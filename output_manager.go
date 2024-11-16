package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/theckman/yacspin"
)

type outputManager struct {
	spinner *yacspin.Spinner
	verbose bool
}

func NewOutputManager(verbose bool) *outputManager {
	return &outputManager{
		verbose: verbose,
	}
}

func config(msg string) yacspin.Config {
	return yacspin.Config{
		CharSet:           yacspin.CharSets[59],
		Frequency:         100 * time.Millisecond,
		Message:           msg,
		ShowCursor:        true,
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✗",
		StopMessage:       msg,
		StopFailMessage:   msg,
		StopFailColors:    []string{"fgRed"},
		SuffixAutoColon:   false,
	}
}
func createSpinner(msg string) *yacspin.Spinner {
	spinner, err := yacspin.New(config(msg))
	if err != nil {
		panic(fmt.Errorf("failed to create spinner: %w", err))
	}
	if !verbose {
		err := spinner.Start()
		if err != nil {
			panic(fmt.Errorf("failed to start spinner: %w", err))
		}
	}
	return spinner
}

func (o *outputManager) info(msg string) {
	if verbose {
		fmt.Println(msg)
	} else {
		if o.spinner != nil {
			o.spinner.Stop()
		}
		o.spinner = createSpinner(msg)
	}
}

func (o *outputManager) debug(msg string) {
	if verbose {
		o.info(msg)
	}
}

func (o *outputManager) error(err error) {
	msg := fmt.Sprintf("Failure: %v", err)
	if verbose {
		fmt.Println(msg)
		fmt.Printf("%s\n", debug.Stack())
	} else {
		o.spinner.StopFail()
		fmt.Println(msg)
	}
}

func (o *outputManager) close(msg string) {
	if verbose {
		fmt.Println(msg)
	} else {
		o.spinner.Stop()
		fmt.Println(msg)
	}
}
