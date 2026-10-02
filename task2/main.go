package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
	
    cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid arguments: %v\n", err)
		os.Exit(1)
	}

	run(ctx, cfg)

    if ctx.Err() != nil && ctx.Err() != io.EOF{
        fmt.Fprintf(os.Stderr, "stopped: %v\n", context.Cause(ctx))
    }
}
