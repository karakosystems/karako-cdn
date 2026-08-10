// Command healthcheck is the Docker HEALTHCHECK probe for the scratch image
// (no shell or curl available). Exits 0 when /health answers 200.
package main

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	if _, err := strconv.Atoi(port); err != nil {
		os.Exit(1)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:" + port + "/health") // #nosec G704 -- localhost probe, port validated above
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
