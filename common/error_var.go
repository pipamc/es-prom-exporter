package common

import "github.com/pkg/errors"

func HTTPStatusError(statusCode int) error {
	return errors.Errorf("HTTP Request failed with code %d", statusCode)
}
