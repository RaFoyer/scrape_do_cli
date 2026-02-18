package errfmt

import (
	"errors"
	"fmt"
	"strings"
)

type wrappedError interface {
	Unwrap() error
}

func Format(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		msg = "unknown error"
	}
	var w wrappedError
	if errors.As(err, &w) {
		inner := strings.TrimSpace(w.Unwrap().Error())
		if inner != "" && inner != msg {
			return fmt.Sprintf("%s: %s", msg, inner)
		}
	}
	return msg
}
