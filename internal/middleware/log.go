package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

type contextKey string

const LoggerKey contextKey = "zlog"

type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader intercepts ResponseWriter.WriteHeader
func (rw *logResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts ResponseWriter.Write
func (rw *logResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

func genShortID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// // RequestLogger record request log and inject Request-ID
// func RequestLogger() func(restful.Handler) restful.Handler {
//     return func(next restful.Handler) restful.Handler {
//         return restful.HandlerFunc(
//             func(w restful.ResponseWriter, r *restful.Request) {
//                 start := time.Now()
//                 // reqID := time.Now().Format("20060102150405.000000")
//                 reqID := genShortID()
//
//                 // child zlog
//                 reqLogger := logger.With(
//                     slog.String("req_id", reqID),
//                 )
//                 ctx := context.WithValue(r.Context(), LoggerKey, reqLogger)
//
//                 // Encapsulated ResponseWriter
//                 mRespWriter := &logResponseWriter{
//                     ResponseWriter: w,
//                     statusCode:     restful.StatusOK, // default 200
//                 }
//                 next.ServeHTTP(mRespWriter, r.WithContext(ctx))
//
//                 duration := time.Since(start)
//                 reqLogger.Info(
//                     "HTTP",
//                     slog.String("method", r.Method),
//                     slog.Int("CODE", mRespWriter.statusCode),
//                     slog.String("URL", r.URL.Path),
//                     // slog.String("ip", r.RemoteAddr),
//                     slog.Int64("TIME", duration.Milliseconds()),
//                 )
//             },
//         )
//     }
// }

// GetLogger gets Logger instance from context，if
// failed, returns default zlog
func GetLogger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
