package main

import (
	"context"
	"fmt"
	"os"

	"auth/internal/server"
)

func main() {
	ctx := context.Background()
	if err := server.Exec(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr, os.Getenv, os.Getwd); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
