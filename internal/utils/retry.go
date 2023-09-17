package utils

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strconv"
	"time"
)

var opErr *net.OpError

var NetworkErrorsToRetry = []error{
	net.ErrWriteToConnected,
	net.ErrClosed,
	http.ErrHandlerTimeout,
	opErr,
}

var defaultIntervals = []int{1, 3, 5}

type RetryFunc func() (interface{}, error)

func RetryFunctionCall(logger *zap.Logger, intervals []int, errorsToRetry []error, fn RetryFunc) (interface{}, error) {
	if intervals == nil {
		intervals = defaultIntervals
	}

	var result interface{}
	var err error
	for _, interval := range intervals {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		if errorsToRetry != nil && !containsError(errorsToRetry, err) {
			return nil, err
		}

		logger.Info("attempt failed with error",
			zap.Error(err),
			zap.String("interval", strconv.Itoa(interval)),
		)
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return nil, fmt.Errorf("max retries reached due to error: %v", err)
}

func containsError(errorsSlice []error, err error) bool {
	for _, e := range errorsSlice {
		if errors.As(err, &e) {
			return true
		}
	}
	return false
}
