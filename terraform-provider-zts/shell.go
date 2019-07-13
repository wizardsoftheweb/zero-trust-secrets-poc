package main

import (
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

// Contains all the useful output from an exec run
type CommandResponse struct {
	// The IO, if any
	result string
	// The exit code
	exitCode int
	// The exit error, if any
	exitErr error
}

// Returns true on 0 exit code
func (c CommandResponse) Succeeded() bool {
	return 0 == c.exitCode
}

// Exposes the result for easy consumption
func (c CommandResponse) String() string {
	return string(c.result)
}

// Processes the thrown error to get the actual exit error
func parseExitCode(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 0
}

// execs a command using CombinedOutputs with some post-processing to
// make it easier to consume
func execCmd(args ...string) CommandResponse {
	logrus.Trace(args)
	process := exec.Command(args[0], args[1:]...)
	combined, err := process.CombinedOutput()
	return CommandResponse{
		result:   string(combined),
		exitCode: parseExitCode(err),
		exitErr:  err,
	}
}
