package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	rootFlag := flag.String("root", "", "repository root (default: walk up to og.config.json)")
	flag.Parse()

	root := *rootFlag
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		root, err = findRoot(wd)
		if err != nil {
			log.Fatal(err)
		}
	}

	cfg, err := LoadConfig(filepath.Join(root, "og.config.json"))
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range cfg.Projects {
		outs, err := Render(root, p, time.Now())
		if err != nil {
			log.Fatalf("project %s: %v", p.Name, err)
		}
		for _, out := range outs {
			fmt.Printf("generated %s\n", out)
		}
	}
}
