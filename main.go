package main

import (
	"os"

  "https://github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
