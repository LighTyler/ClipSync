package main

import (
	"log"

	"golang.design/x/clipboard"

	"github.com/LighTyler/ClipSync/config"
)

func main() {
	log.Printf("WeAreClipboard backend starting...")

	err := clipboard.Init()
	if err != nil {
		log.Fatalf("main - Clipboard init: %v", err)
	}
	log.Printf("main - Clipboard init: success")

	config.NewConfig("")

}
