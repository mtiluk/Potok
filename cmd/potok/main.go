package main

import (
	"fmt"
	"log"
	"os"

	"github.com/michaeltukdev/Potok/internal/client/cli"
	"github.com/michaeltukdev/Potok/internal/client/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.Execute(&cli.Env{
		Config: cfg,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "potok:", err)
		os.Exit(1)
	}
}
