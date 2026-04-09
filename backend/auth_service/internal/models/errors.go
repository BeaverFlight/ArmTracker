package models

import (
	"errors"
)

var (
	ErrUnknown = errors.New("неизвестная ошибка")
)

var knownErrors = []struct {
	err  error
	code int
}{}

func HTTPCode(err error) int {
	for _, e := range knownErrors {
		if errors.Is(err, e.err) {
			return e.code
		}
	}
	return 0
}
