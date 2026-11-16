package crash

import (
	"errors"
	"os"
	"runtime/debug"
)

var ErrNoCrash = errors.New("crash: log not found")

func Get(file string) ([]byte, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, ErrNoCrash
	}
	if info.Size() <= 0 {
		return nil, ErrNoCrash
	}
	return os.ReadFile(file)
}

func Collect(file string) error {
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	return debug.SetCrashOutput(fd, debug.CrashOptions{})
}
