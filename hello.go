package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var urlStore = make(map[string]string)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// ------------------- CORS -------------------
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// ------------------- UTIL -------------------
func generateShortCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// ------------------- STRUCTS -------------------
type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"short_url"`
}

// ------------------- HANDLERS -------------------
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)

	// Allow only POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic URL validation
	if req.URL == "" || !strings.HasPrefix(req.URL, "http") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Generate unique short code
	var shortCode string
	for {
		shortCode = generateShortCode(6)
		if _, exists := urlStore[shortCode]; !exists {
			break
		}
	}

	// Store mapping
	urlStore[shortCode] = req.URL

	resp := Response{
		ShortURL: fmt.Sprintf("https://url-shortener.onrender.com/%s", shortCode),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)

	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	originalURL, exists := urlStore[shortCode]
	if !exists {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// ------------------- MAIN -------------------
func main() {
	// Seed random generator
	rand.Seed(time.Now().UnixNano())

	// Get PORT (Render requirement)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Routes
	http.HandleFunc("/shorten", shortenHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
		redirectHandler(w, r)
	})

	fmt.Println("Server running on port:", port)
	http.ListenAndServe(":"+port, nil)
}
