package chain

import "fmt"

type closeStack []func() error

func (stack closeStack) Close() error {
	var errors []error

	for i := len(stack) - 1; i >= 0; i-- {
		err := stack[i]()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return CloseErrorStack(errors)
	}

	return nil
}

type CloseErrorStack []error

func (ces CloseErrorStack) Error() string {
	return fmt.Sprint([]error(ces))
}
