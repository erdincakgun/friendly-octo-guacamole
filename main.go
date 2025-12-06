package main

import (
	"context"
	"encoding/json"
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
)

// MenuItem represents a food item in our delivery platform
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

// Server holds our application state
type Server struct {
	menuItems map[string]MenuItem
}

// NewServer creates a new server instance with sample menu items
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
	w.Write(data)

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
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", http.StatusOK).
		Dur("duration", time.Since(start)).
		Msg("Health check successful")
}

func (s *Server) menuHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 10% chance of simulated failure for testing monitoring
	if rand.Float32() < 0.1 {
		status := http.StatusInternalServerError
		writeError(w, status, "Failed to fetch menu items from restaurant database")

		log.Error().
			Err(fmt.Errorf("restaurant service timeout")).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Dur("duration", time.Since(start)).
			Msg("Failed to list menu items")

		return
	}

	menuList := make([]MenuItem, 0, len(s.menuItems))
	for _, item := range s.menuItems {
		menuList = append(menuList, item)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"menu_items": menuList,
		"count":      len(menuList),
	})

	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", http.StatusOK).
		Dur("duration", time.Since(start)).
		Int("count", len(menuList)).
		Msg("Listed menu items")
}

func (s *Server) menuItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	menuItemID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/menu/"))

	if menuItemID == "" {
		status := http.StatusBadRequest
		writeError(w, status, "Menu item ID is required")

		log.Warn().
			Err(fmt.Errorf("missing menu item ID")).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Dur("duration", time.Since(start)).
			Msg("Invalid menu item request")
		return
	}

	menuItem, exists := s.menuItems[menuItemID]
	if !exists {
		status := http.StatusNotFound
		writeError(w, status, fmt.Sprintf("Menu item with ID '%s' not found", menuItemID))

		log.Warn().
			Err(fmt.Errorf("menu item not found")).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Dur("duration", time.Since(start)).
			Str("menu_item_id", menuItemID).
			Msg("Menu item ID does not exist")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"menu_item": menuItem,
	})

	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", http.StatusOK).
		Dur("duration", time.Since(start)).
		Str("menu_item_id", menuItemID).
		Str("menu_item_name", menuItem.Name).
		Str("restaurant", menuItem.Restaurant).
		Msg("Retrieved menu item")
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/api/menu", server.menuHandler)
	mux.HandleFunc("/api/menu/", server.menuItemByIDHandler)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
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

	log.Info().Msg("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}
