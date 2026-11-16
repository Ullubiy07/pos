package main

import (
	"context"
	"fmt"
	"os"

	"pos/internal/di"
)

func main() {
	container := di.New()
	app := container.CLI(os.Stdin, os.Stdout)
	if err := app.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
