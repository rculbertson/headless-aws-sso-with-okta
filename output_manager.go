package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/theckman/yacspin"
)

type outputManager struct {
	spinner    *yacspin.Spinner
	lastStatus string
	verbose    bool
}

func NewOutputManager(verbose bool) *outputManager {
	cfg := yacspin.Config{
		CharSet:           yacspin.CharSets[59],
		Frequency:         100 * time.Millisecond,
		Message:           "Initializing",
		ShowCursor:        true,
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✗",
		StopMessage:       "Logged in successfully",
		StopFailMessage:   "Log in failed",
		StopFailColors:    []string{"fgRed"},
		Suffix:            "AWS SSO: ",
		SuffixAutoColon:   false,
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create spinner: %w", err))
	}
	if !verbose {
		err := spinner.Start()
		if err != nil {
			panic(fmt.Errorf("failed to start spinner: %w", err))
		}
	}
	x := &outputManager{
		lastStatus: "Iinitializing",
		spinner:    spinner,
		verbose:    verbose,
	}
	return x
}

func (o *outputManager) info(msg string) {
	o.lastStatus = msg
	if verbose {
		fmt.Println(msg)
	} else {
		o.spinner.Message(msg)
	}
}

func (o *outputManager) debug(msg string) {
	if verbose {
		o.info(msg)
	}
}

func (o *outputManager) error(err error) {
	msg := fmt.Sprintf("Failure while %s: %v", strings.ToLower(o.lastStatus), err)
	if verbose {
		fmt.Println(msg)
		fmt.Printf("%s\n", debug.Stack())
	} else {
		red := color.New(color.FgRed).SprintFunc()
		o.spinner.StopFailMessage(red(msg))
		o.spinner.StopFail()
	}
}

func (o *outputManager) close(msg string) {
	if verbose {
		fmt.Println(msg)
	} else {
		o.spinner.StopMessage(msg)
		o.spinner.Stop()
	}
}
