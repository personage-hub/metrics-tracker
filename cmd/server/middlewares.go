package main

import (
	"bytes"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

type (
	responseData struct {
		status int
		size   int
		body   []byte
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.body = b
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func requestWithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method := r.Method
		path := r.URL.Path

		requestBodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		logWriter := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&logWriter, r)

		duration := time.Since(start)
		logger.Log.Info("got incoming HTTP request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("request body", string(requestBodyBytes)),
			zap.String("response body", string(responseData.body)),
			zap.String("duration", duration.String()),
			zap.String("size", strconv.Itoa(responseData.size)),
			zap.String("status", strconv.Itoa(responseData.status)),
		)
	})
}
