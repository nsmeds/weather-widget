package main

import (
	"context"
	"fmt"
	"io"
	"os"
)

func Run(ctx context.Context, cancel context.CancelFunc, args []string, stdout, stderr io.Writer) error {
	defer cancel()

	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	if err := Run(ctx, cancel, os.Args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
