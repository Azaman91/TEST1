package main

import (
	rest "Project/REST"
	_ "Project/docs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func main() {
	if err := rest.InitDB(); err != nil {
		log.Fatal("Error DB")
	}

	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger.yaml")))
	mux.Handle("/", loggingMiddleware(logger, http.HandlerFunc(rest.SubHandler)))
	mux.Handle("/read", loggingMiddleware(logger, http.HandlerFunc(rest.ListSubscriptionHandler)))
	mux.Handle("/read/{id}", loggingMiddleware(logger, http.HandlerFunc(rest.GetSubscriptionHandler)))
	mux.Handle("/update", loggingMiddleware(logger, http.HandlerFunc(rest.UpdateSubscriptionHandler)))
	mux.Handle("/delete", loggingMiddleware(logger, http.HandlerFunc(rest.DeleteSubscriptionHandler)))
	mux.Handle("/total", loggingMiddleware(logger, http.HandlerFunc(rest.TotalCostHandler)))
	http.ListenAndServe(":8080", mux)
}
