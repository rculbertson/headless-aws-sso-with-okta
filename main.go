package main

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var (
	scanner         = bufio.NewScanner(os.Stdin)
	screenshotCount = 0
	sessionId       = strconv.FormatInt(time.Now().Unix(), 10)
	version         = "dev"
	showBrowser     bool
	verbose         bool
	captureState    bool
	oktaAuth        string
	anchorLabel     string
	email           string
	out             *outputManager
)

const cookieFileName = ".headless-aws-sso-with-okta"

func main() {
	flag.BoolVar(&showBrowser, "show-browser", false, "Show browser window during login")
	flag.BoolVar(&verbose, "verbose", false, "Print verbose output")
	flag.BoolVar(&captureState, "capture-state", false, "Take screenshots and dump html of each login page")
	flag.StringVar(&oktaAuth, "okta-auth", "push-notification", "Okta authentication method (fastpass or push-notification)")
	flag.StringVar(&email, "email", "", "email to sign in with. Okta FastPass will be used if not specified.)")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	out = NewOutputManager(verbose)
	out.info("Initializing")

	flag.VisitAll(func(f *flag.Flag) {
		out.debug(fmt.Sprintf("%s=%v", f.Name, f.Value))
	})

	defer func() {
		if r := recover(); r != nil {
			out.error(fmt.Errorf("%v", r))
			os.Exit(1)
		}
	}()

	if oktaAuth != "fastpass" && oktaAuth != "push-notification" {
		panic("invalid value for okta-auth flag: must be either 'fastpass' or 'push-notification'")
	}

	var label string
	if oktaAuth == "push-notification" {
		label = "Select to get a push notification to the Okta Verify app."
	} else {
		label = "Select Okta FastPass."
	}
	anchorLabel = fmt.Sprintf("a[aria-label='%s']", label)
	url := getURL()

	login(url)

	// Consume and display message saying login was successful
	out.debug("Waiting to receive success confirmation")
	line := readLineFromStdin(time.Second * 10)
	out.close(line)
}

func readLineFromStdin(timeout time.Duration) string {
	lineChan := make(chan string)
	errChan := make(chan error)

	go func() {
		if scanner.Scan() {
			line := scanner.Text()
			if verbose {
				fmt.Printf("STDIN: %s\n", line)
			}
			lineChan <- line
		} else if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	select {
	case line := <-lineChan:
		return line
	case err := <-errChan:
		panic(err)
	case <-time.After(timeout):
		panic(fmt.Errorf("timed out"))
	}
}

func getURL() string {
	out.debug("Reading url from stdin")
	for {
		line := readLineFromStdin(time.Second * 5)
		pattern, _ := regexp.Compile("^https.*user_code=([A-Z]{4}-?){2}")
		if pattern.MatchString(line) {
			return line
		}
	}
}

func click(page *rod.Page, elem *rod.Element) {
	capturePageState(page)
	elem.MustWaitEnabled().MustClick()
	page.MustWaitLoad()
	capturePageState(page)
}

func handleSignIn(page *rod.Page, fastPassButton *rod.Element) {
	// Sign in with email if one was provided, otherwise open okta verify
	if email != "" {
		out.debug("Submitting email " + email)
		page.MustElement(`input[name="identifier"]`).MustInput(email)
		nextButton := page.MustElement(`input[value="Next"]`)
		click(page, nextButton)
	} else {
		out.debug("Opening Okta Verify")
		click(page, fastPassButton)
	}

	page.Race().ElementR("a", "Open Okta Verify").MustHandle(func(elem *rod.Element) {
		panic("Okta Verify is not installed. Please sign in with email.")
	}).ElementR(anchorLabel, "Select").MustHandle(func(selectElem *rod.Element) {
		out.debug(fmt.Sprintf("Selecting to authenticate with %s", oktaAuth))
		click(page, selectElem)
		out.info("Waiting for user to authenticate")
		allowElem := page.MustElementR("button", "Allow")
		handleAllow(page, allowElem)
	}).MustDo()

}

func handleAllow(page *rod.Page, elem *rod.Element) {
	click(page, elem)
	out.info("Waiting for authentication to complete")
	// After clicking "Allow", we must wait for the "Request approved" screen to appear
	// before closing the browser, otherwise the login will not complete.
	page.MustElementR("div", "Request approved")
}

func login(url string) {
	out.debug("Initializing browser")
	var browser *rod.Browser
	if !showBrowser {
		browser = rod.New().MustConnect()
	} else {
		url := launcher.New().Headless(false).MustLaunch()
		browser = rod.New().ControlURL(url).MustConnect()
	}
	defer browser.MustClose()
	out.info("Submitting request to Okta")
	loadCookies(*browser)
	out.debug("Loading " + url)
	page := browser.MustPage(url).Timeout(time.Minute * 2).MustWaitLoad()
	capturePageState(page)
	// Authorization requested page with confirmation code
	elem := page.MustElementR("button", "Confirm and continue")
	click(page, elem)

	page.Race().ElementR("a", "Sign in with Okta FastPass").MustHandle(func(elem *rod.Element) {
		// Okta sign in page, with "Sign in with Okta FastPass" button, and a username input textbox
		handleSignIn(page, elem)
	}).ElementR("button", "Allow").MustHandle(func(elem *rod.Element) {
		// If the user has saved cookies, we'll jump right to the "Allow Access" screen.
		handleAllow(page, elem)
	}).MustDo()

	saveCookies(*browser)
}

func capturePageState(page *rod.Page) {
	if captureState {
		html, err := page.HTML()
		if err != nil {
			panic(err)
		}
		os.WriteFile(fmt.Sprintf("screen-%s-%d.html", sessionId, screenshotCount), []byte(html), 0644)
		page.MustScreenshotFullPage(fmt.Sprintf("sso-screenshot-%s-%d.png", sessionId, screenshotCount))
		screenshotCount++
	}
}
func loadCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dirname, cookieFileName)
	out.debug("Loading cookies: " + path)
	data, _ := os.ReadFile(dirname + "/.headless-aws-sso-with-okta")
	sEnc, _ := b64.StdEncoding.DecodeString(string(data))
	var cookie *proto.NetworkCookie
	json.Unmarshal(sEnc, &cookie)

	if cookie != nil {
		browser.MustSetCookies(cookie)
	}
}

func saveCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dirname, cookieFileName)
	out.debug("Saving cookies: " + path)
	cookies := (browser.MustGetCookies())

	for _, cookie := range cookies {
		if cookie.Name == "x-amz-sso_authn" {
			data, _ := json.Marshal(cookie)

			sEnc := b64.StdEncoding.EncodeToString([]byte(data))
			err = os.WriteFile(path, []byte(sEnc), 0644)

			if err != nil {
				panic(err)
			}
			break
		}
	}
}
