package utils

import "fmt"

type ConnectionError struct {
	Code    string
	Message string
	Details error
}

func (e *ConnectionError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
