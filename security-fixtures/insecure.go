// Package fixtures contains intentionally unsafe code used only to exercise CI scanners.
package fixtures

import "os/exec"

const fakeAPIKey = "AKIA1234567890ABCDEF"

// RunShellCommand is never called by the application. It exists to exercise SAST rules.
func RunShellCommand(userInput string) *exec.Cmd {
	return exec.Command("sh", "-c", userInput)
}

func FakeAPIKey() string {
	return fakeAPIKey
}
