// main.go
package main

import (
	"log"
	"net/http"
)

func main() {
	// Serve index.html at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Serve the CSS file at /styles.css
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "styles.css")
	})
	// Serve the video file at /aerial.mp4
	http.HandleFunc("/aerial.mp4", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/video/aerial.mp4")
	})

	// Start server on port 8080
	log.Println("Server listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
