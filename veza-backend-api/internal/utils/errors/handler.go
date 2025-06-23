// internal/utils/errors/handler.go
package errors

import (
    "encoding/json"
    "log/slog"
    "net/http"
)

type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
    TraceID string `json:"trace_id,omitempty"`
}

func (e APIError) Error() string {
    return e.Message
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
    traceID := r.Header.Get("X-Trace-ID")
    
    apiErr := APIError{
        Code:    statusCode,
        Message: err.Error(),
        TraceID: traceID,
    }
    
    slog.Error("API Error",
        "error", err.Error(),
        "status_code", statusCode,
        "path", r.URL.Path,
        "method", r.Method,
        "trace_id", traceID,
    )
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(apiErr)
}