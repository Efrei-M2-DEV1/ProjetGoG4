package main

import (
	_ "github.com/axellelanca/urlshortener/cmd/cli"    // Importe le package 'cli' pour que ses init() soient exécutés
	_ "github.com/axellelanca/urlshortener/cmd/server" // Importe le package 'server' pour que ses init() soient exécutés
	"github.com/axellelanca/urlshortener/cmd"
)

func main() {
	cmd.Execute()
}
