package main

import "errors"

func errorsIs(err, target error) bool {
	return errors.Is(err, target)
}
