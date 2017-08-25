package logging

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger
var loggerLock sync.Mutex

func Logger() *zap.Logger {
	loggerLock.Lock()
	defer loggerLock.Unlock()
	if logger == nil {
		var err error
		if os.Getenv("ENV") == "production" {
			conf := zap.NewProductionConfig()
			conf.EncoderConfig.MessageKey = "log"
			logger, err = conf.Build()
		} else {
			logger, err = zap.NewDevelopment()
		}
		if err != nil {
			panic(err)
		}
	}
	return logger
}

// Returns a logger with the request id from the context (x-fair-request-id).
// If no request id exists the field is not added.
func WithRequestId(ctx context.Context) *zap.Logger {
	log := Logger()
	if ctx != nil {
		if requestId, ok := ctx.Value("x-fair-request-id").(string); ok {
			if requestId != "" {
				return log.With(zap.String("requestId", requestId))
			}
		}
	}
	return log
}

type responseWriter struct {
	status int
	rw     http.ResponseWriter
}

func (rw *responseWriter) Header() http.Header {
	return rw.rw.Header()
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.rw.Write(b)
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.rw.WriteHeader(status)
}

// A response writer that impliments the Hijacker interface
type responseWriterHijacker struct {
	responseWriter
	hijacker http.Hijacker
}

func (rw *responseWriterHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.hijacker.Hijack()
}

// Given a zap logger and an http handler, returns a handler that logs the request
// start and request end.
func Wrap(handler http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log := WithRequestId(r.Context())
		start := time.Now()
		requestURI := r.URL.RequestURI()
		log.Info("request recieved",
			zap.String("method", r.Method),
			zap.String("uri", requestURI),
		)
		rwWrapper := responseWriter{rw: rw}
		if hijacker, ok := rw.(http.Hijacker); ok {
			handler.ServeHTTP(&responseWriterHijacker{rwWrapper, hijacker}, r)
		} else {
			handler.ServeHTTP(&rwWrapper, r)
		}
		log.Info("request finished",
			zap.String("method", r.Method),
			zap.String("uri", requestURI),
			zap.Int("status", rwWrapper.status),
			zap.Duration("time", time.Since(start)),
		)
	}
}
