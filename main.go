package main

import (
	"bufio"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/theckman/yacspin"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Time before MFA step times out
const MFA_TIMEOUT = 30

var cfg = yacspin.Config{
	Frequency:         100 * time.Millisecond,
	CharSet:           yacspin.CharSets[59],
	Suffix:            "AWS SSO Signing in: ",
	SuffixAutoColon:   false,
	Message:           "",
	StopCharacter:     "✓",
	StopFailCharacter: "✗",
	StopMessage:       "Logged in successfully",
	StopFailMessage:   "Log in failed",
	StopColors:        []string{"fgGreen"},
}

var spinner, _ = yacspin.New(cfg)

type Credential struct {
	Email    string
	Login    string
	Password string
}

func main() {
	spinner.Start()
	mfa_code := os.Args[1]

	// get sso url from stdin
	url := getURL()
	// start aws sso login
	ssoLogin(mfa_code, url)

	spinner.Stop()
	time.Sleep(1 * time.Second)
}

// returns sso url from stdin.
func getURL() string {
	spinner.Message("reading url from stdin")

	scanner := bufio.NewScanner(os.Stdin)
	url := ""
	for url == "" {
		scanner.Scan()
		t := scanner.Text()
		r, _ := regexp.Compile("^https.*user_code=([A-Z]{4}-?){2}")

		if r.MatchString(t) {
			url = t
		}
	}

	return url
}

// get okta credentials from dashlane
func getCredentials() (string, string) {
	spinner.Message("fetching credentials from Dashlane")

	// Run dcli from shell, receive output in JSON format
	cmd := exec.Command("bin/dcli", "password", "okta", "--output", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	// Extract okta creds from output JSON
	var arr []Credential
	_ = json.Unmarshal(out, &arr)
	creds := arr[0]

	var username string
	if len(creds.Email) > 0 {
		username = creds.Email
	} else if len(creds.Login) > 0 {
		username = creds.Login
	} else {
		error("Cannot find a valid login.")
	}
	passphrase := creds.Password
	return username, passphrase
}

// login with hardware MFA
func ssoLogin(mfa_code string, url string) {
	username, passphrase := getCredentials()
	spinner.Message(color.MagentaString("init headless-browser"))
	spinner.Pause()
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	err := rod.Try(func() {
		page := browser.MustPage(url)

		// authorize
		spinner.Unpause()
		spinner.Message("logging in")
		page.MustElementR("button", "Next").MustWaitEnabled().MustClick()

		// sign-in
		oktaSignIn(*page, username, passphrase)
		oktaAuthMfa(*page, mfa_code)

		// allow request
		unauthorized := true
		for unauthorized {

			txt := page.Timeout(MFA_TIMEOUT * time.Second).MustElement(".awsui-util-mb-s").MustWaitLoad().MustText()
			if txt == "Request approved" {
				unauthorized = false
			} else {
				exists, _, _ := page.HasR("button", "Allow")
				if exists {
					page.MustWaitLoad().MustElementR("button", "Allow").MustClick()
				}

				time.Sleep(500 * time.Millisecond)
			}
		}
	})

	if errors.Is(err, context.DeadlineExceeded) {
		panic("Timed out waiting for MFA")
	} else if err != nil {
		panic(err.Error())
	}
}

// executes okta signin step
func oktaSignIn(page rod.Page, username, passphrase string) {
	page.Timeout(MFA_TIMEOUT * time.Second).MustElement(".okta-sign-in-header").MustWaitLoad()
	page.MustElement("#okta-signin-username").MustInput(username)
	page.MustElement("#okta-signin-password").MustInput(passphrase)
	page.MustWaitLoad().MustElementR("input", "Sign In").MustClick()
}

// TODO: allow user to enter MFA Code
func mfa(page rod.Page) {
	_ = beeep.Notify("headless-sso", "Touch U2F device to proceed with authenticating AWS SSO", "")
	_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)

	spinner.Message(color.YellowString("Touch U2F"))
}

func oktaAuthMfa(page rod.Page, mfaCode string) {
	page.MustElement("form.mfa-verify-totp input").MustInput(mfaCode)
	page.MustWaitLoad().MustElementR("input", "Verify").MustClick()
}

// load cookies
func loadCookies(browser rod.Browser) {
	spinner.Message("loading cookies")
	dirname, err := os.UserHomeDir()
	if err != nil {
		error(err.Error())
	}

	data, _ := os.ReadFile(dirname + "/.headless-sso")
	sEnc, _ := b64.StdEncoding.DecodeString(string(data))
	var cookie *proto.NetworkCookie
	json.Unmarshal(sEnc, &cookie)

	if cookie != nil {
		browser.MustSetCookies(cookie)
	}
}

// save authn cookie
func saveCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		error(err.Error())
	}

	cookies := (browser.MustGetCookies())

	for _, cookie := range cookies {
		if cookie.Name == "x-amz-sso_authn" {
			data, _ := json.Marshal(cookie)

			sEnc := b64.StdEncoding.EncodeToString([]byte(data))
			err = os.WriteFile(dirname+"/.headless-sso", []byte(sEnc), 0644)

			if err != nil {
				error("Failed to save x-amz-sso_authn cookie")
			}
			break
		}
	}
}

// print error message and exit
func panic(errorMsg string) {
	red := color.New(color.FgRed).SprintFunc()
	spinner.StopFailMessage(red("Login failed error - " + errorMsg))
	spinner.StopFail()
	os.Exit(1)
}

// print error message
func error(errorMsg string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	spinner.Message("Warn: " + yellow(errorMsg))
}
