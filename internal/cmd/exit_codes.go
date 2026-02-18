package cmd

import (
	"context"
	"errors"
	"net"
	"net/url"

	"github.com/ra/scrape_do_cli/internal/client"
)

const (
	exitCodeAuthRequired = 4
	exitCodeRateLimited  = 7
	exitCodeRetryable    = 8
	exitCodeConfig       = 10
	exitCodeCancelled    = 130
)

func stableExitCode(err error) error {
	if err == nil {
		return nil
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return err
	}

	if errors.Is(err, context.Canceled) {
		return &ExitError{Code: exitCodeCancelled, Err: err}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &ExitError{Code: exitCodeRetryable, Err: err}
	}

	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401, 403:
			return &ExitError{Code: exitCodeAuthRequired, Err: err}
		case 429:
			return &ExitError{Code: exitCodeRateLimited, Err: err}
		default:
			if apiErr.StatusCode >= 500 {
				return &ExitError{Code: exitCodeRetryable, Err: err}
			}
		}
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return &ExitError{Code: exitCodeRetryable, Err: err}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &ExitError{Code: exitCodeRetryable, Err: err}
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var dErr *net.DNSError
		if errors.As(urlErr.Err, &dErr) {
			return &ExitError{Code: exitCodeRetryable, Err: err}
		}
		var nErr net.Error
		if errors.As(urlErr.Err, &nErr) && nErr.Timeout() {
			return &ExitError{Code: exitCodeRetryable, Err: err}
		}
	}

	return err
}
