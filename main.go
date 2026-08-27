package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	dirs := flag.String("d", ".", "dir/path to serve from")
	port := flag.Int("p", 8000, "port to listen on")
	bind := flag.String("b", "127.0.0.1", "address to bind to")

	flag.Parse()

	if _, err := os.Stat(*dirs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: directory %q does not exist\n", *dirs)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", *bind, *port)

	handler := &server{dir: *dirs}

	abs, err := filepath.Abs(*dirs)
	if err != nil {
		abs = *dirs
	}

	fmt.Printf("Serving: %s\n", abs)
	fmt.Printf("Listening on http://%s\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
