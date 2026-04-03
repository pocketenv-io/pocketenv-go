package pocketenv

import "fmt"

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("pocketenv: HTTP %d: %s", e.StatusCode, e.Message)
}
