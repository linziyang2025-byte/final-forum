package api

import "errors"

func ErrorsIs(err, target error) bool {
	return errors.Is(err, target)
}
