package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// LoggingResponseWriter wraps http.ResponseWriter to capture status code
type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// logToFile logs a message to server.log with a timestamp
func logToFile(message string) {
	f, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file:", err)
		return
	}
	defer f.Close()
	logLine := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), message)
	f.WriteString(logLine)
}

// loggingMiddleware logs each request and response
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &LoggingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for downstream

		start := time.Now()
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)

		logMsg := fmt.Sprintf("Method: %s, Path: %s, Body: %s, Status: %d, Duration: %s",
			r.Method, r.URL.Path, string(body), lrw.StatusCode, duration)
		logToFile(logMsg)
	})
}
