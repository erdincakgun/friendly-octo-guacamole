package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

func logJSON(level, method, path string, status int, duration time.Duration, msg string, err error, menuItemID string) {
	entry := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"level":       level,
		"method":      method,
		"path":        path,
		"status_code": status,
		"duration_ms": duration.Seconds() * 1000,
		"message":     msg,
	}

	if err != nil {
		entry["error"] = err.Error()
	}
	if menuItemID != "" {
		entry["menu_item_id"] = menuItemID
	}

	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	logJSON("INFO", r.Method, r.URL.Path, http.StatusOK, time.Since(start),
		"Health check successful", nil, "")
}

func (s *Server) menuHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 10% chance of simulated failure for testing monitoring
	if rand.Float32() < 0.1 {
		status := http.StatusInternalServerError
		writeError(w, status, "Failed to fetch menu items from restaurant database")

		logJSON("ERROR", r.Method, r.URL.Path, status, time.Since(start),
			"Failed to list menu items", fmt.Errorf("restaurant service timeout"), "")
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

	logJSON("INFO", r.Method, r.URL.Path, http.StatusOK, time.Since(start),
		fmt.Sprintf("Listed %d menu items", len(menuList)), nil, "")
}

func (s *Server) menuItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	menuItemID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/menu/"))

	if menuItemID == "" {
		status := http.StatusBadRequest
		writeError(w, status, "Menu item ID is required")

		logJSON("WARN", r.Method, r.URL.Path, status, time.Since(start),
			"Invalid menu item request", fmt.Errorf("missing menu item ID"), "")
		return
	}

	menuItem, exists := s.menuItems[menuItemID]
	if !exists {
		status := http.StatusNotFound
		writeError(w, status, fmt.Sprintf("Menu item with ID '%s' not found", menuItemID))

		logJSON("WARN", r.Method, r.URL.Path, status, time.Since(start),
			fmt.Sprintf("Menu item ID %s does not exist", menuItemID),
			fmt.Errorf("menu item not found"), menuItemID)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"menu_item": menuItem,
	})

	logJSON("INFO", r.Method, r.URL.Path, http.StatusOK, time.Since(start),
		fmt.Sprintf("Retrieved menu item: %s from %s", menuItem.Name, menuItem.Restaurant),
		nil, menuItemID)
}

func main() {
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

	logJSON("INFO", "", "", 0, 0, "Starting server on :8080", nil, "")

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logJSON("INFO", "", "", 0, 0, "Shutting down server...", nil, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logJSON("ERROR", "", "", 0, 0, "Server forced to shutdown", err, "")
		log.Fatal("Server forced to shutdown:", err)
	}

	logJSON("INFO", "", "", 0, 0, "Server exited gracefully", nil, "")
}
