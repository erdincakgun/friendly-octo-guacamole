package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var tracer = otel.Tracer("github.com/blackswan/mock-go")

func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("friendly-octo-guacamole"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		handleErr(err)
		return
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		handleErr(err)
		return
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return
}

type MenuItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
	Description string  `json:"description"`
	Restaurant  string  `json:"restaurant"`
	Category    string  `json:"category"`
	PrepTime    int     `json:"prep_time_minutes"`
}

type Server struct {
	menuItems map[string]MenuItem
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		logger := log.Info()
		if rw.statusCode >= 500 {
			logger = log.Error()
		} else if rw.statusCode >= 400 {
			logger = log.Warn()
		}

		logger.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Dur("duration", time.Since(start)).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("HTTP request")
	})
}

func NewServer() *Server {
	return &Server{
		menuItems: map[string]MenuItem{
			"1": {ID: "1", Name: "Margherita Pizza", Price: 12.99, Available: true, Description: "Fresh mozzarella, tomato sauce, basil", Restaurant: "Tony's Pizza", Category: "Pizza", PrepTime: 20},
			"2": {ID: "2", Name: "Chicken Pad Thai", Price: 14.99, Available: true, Description: "Rice noodles, chicken, peanuts, lime", Restaurant: "Thai Palace", Category: "Asian", PrepTime: 15},
			"3": {ID: "3", Name: "Classic Burger", Price: 11.99, Available: false, Description: "Beef patty, lettuce, tomato, cheese", Restaurant: "Burger Joint", Category: "Burgers", PrepTime: 12},
			"4": {ID: "4", Name: "Caesar Salad", Price: 8.99, Available: true, Description: "Romaine lettuce, parmesan, croutons", Restaurant: "Healthy Bites", Category: "Salads", PrepTime: 5},
			"5": {ID: "5", Name: "Sushi Platter", Price: 24.99, Available: true, Description: "12 piece mixed sushi selection", Restaurant: "Sakura Sushi", Category: "Japanese", PrepTime: 25},
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		log.Error().
			Err(err).
			Int("status", status).
			Msg("Failed to marshal JSON response")

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		log.Error().Err(err).Msg("Failed to write response")
		return err
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	errorResponse := map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	}

	data, err := json.Marshal(errorResponse)
	if err != nil {
		log.Error().
			Err(err).
			Int("status", status).
			Str("original_message", message).
			Msg("Failed to marshal error response, using plain text fallback")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	_ = writeJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) menuHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "fetchMenuItems")
	defer span.End()

	if rand.Float32() < 0.1 {
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(errors.New("database connection failed"))
		writeError(w, http.StatusInternalServerError, "Failed to fetch menu items from restaurant database")
		return
	}

	menuList := make([]MenuItem, 0, len(s.menuItems))
	for _, item := range s.menuItems {
		menuList = append(menuList, item)
	}

	span.SetAttributes(attribute.Int("menu.count", len(menuList)))
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"menu_items": menuList,
		"count":      len(menuList),
	})
}

func (s *Server) menuItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "fetchMenuItemByID")
	defer span.End()

	menuItemID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/menu/"))
	span.SetAttributes(attribute.String("menu.item.id", menuItemID))

	if menuItemID == "" {
		span.SetAttributes(attribute.Bool("error", true))
		writeError(w, http.StatusBadRequest, "Menu item ID is required")
		return
	}

	menuItem, exists := s.menuItems[menuItemID]
	if !exists {
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(fmt.Errorf("menu item not found: %s", menuItemID))
		writeError(w, http.StatusNotFound, fmt.Sprintf("Menu item with ID '%s' not found", menuItemID))
		return
	}

	span.SetAttributes(
		attribute.String("menu.item.name", menuItem.Name),
		attribute.Float64("menu.item.price", menuItem.Price),
	)
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"menu_item": menuItem,
	})
}

func newHTTPHandler(server *Server) http.Handler {
	mux := http.NewServeMux()

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc("/health", server.healthHandler)
	handleFunc("/api/menu", server.menuHandler)
	handleFunc("/api/menu/", server.menuItemByIDHandler)

	return otelhttp.NewHandler(mux, "/")
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ctx := context.Background()
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup OpenTelemetry SDK")
	}
	defer func() {
		if err := otelShutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown OpenTelemetry SDK")
		}
	}()

	server := NewServer()
	handler := newHTTPHandler(server)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info().Msgf("Starting server on %s", httpServer.Addr)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}
