package os

import (
	"fmt"

	"gitlab.com/CoiaPrant/exec"
)

func Exec(command string) error {
	return exec.Default.Exec(command)
}

func Execf(format string, a ...any) error {
	command := fmt.Sprintf(format, a...)
	return Exec(command)
}
