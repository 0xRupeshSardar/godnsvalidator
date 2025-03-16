package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xRupeshSardar/godnsvalidator/internal/config"
	"github.com/0xRupeshSardar/godnsvalidator/internal/output"
	"github.com/0xRupeshSardar/godnsvalidator/internal/resolver"
	"github.com/0xRupeshSardar/godnsvalidator/internal/validator"
)

func main() {
	cfg := config.ParseFlags()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupInterrupt(cancel)
	output.Init(cfg)

	baseline := resolver.GetBaseline(cfg)
	if baseline == nil {
		output.Error("Failed to establish baseline")
		os.Exit(1)
	}

	validator.ValidateServers(ctx, cfg, baseline)
	output.WriteResults(cfg)

	if !cfg.Silent {
		output.Success("\nValid servers found: %d", len(validator.ValidServers))
	}
}

func setupInterrupt(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()
}