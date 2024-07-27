package logger

import "net/http"

type (
	responseData struct {
		status   int
		size     int
		location string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseData.location = r.Header().Get("Location")
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
