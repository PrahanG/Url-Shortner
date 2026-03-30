package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var urlSt = make(map[string]string)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func shortcode(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"short_url"`
}

func shorten(w http.ResponseWriter, r *http.Request) {
	var req Request
	json.NewDecoder(r.Body).Decode(&req)
	short := shortcode(6)

	urlSt[short] = req.URL

	resp := Response{
		ShortURL: fmt.Sprintf("http://localhost:8080/%s", short),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	short := r.URL.Path[1:]
	orig, exists := urlSt[short]
	if !exists {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, orig, http.StatusFound)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/shorten", shorten)

	// Serve homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			homePage(w, r)
			return
		}
		redirect(w, r)
	})

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)

}
