package chain

import (
	"errors"
	"fmt"
	"io"
)

type closeStack []io.Closer

func (stack closeStack) Close() error {
	var errors []error

	for i := len(stack) - 1; i >= 0; i-- {
		err := stack[i].Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return closeErrorStack(errors)
	}

	return nil
}

type closeErrorStack []error

func (ces closeErrorStack) Error() string {
	return fmt.Sprint([]error(ces))
}

func (ces closeErrorStack) Is(target error) bool {
	for _, err := range ces {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (ces closeErrorStack) As(target interface{}) bool {
	for _, err := range ces {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}
