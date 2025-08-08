// Example HTTP server for testing the multiminer library.
// This is NOT part of the core library - it's just an example implementation
// showing how to use the library in a server application.
//
// Library consumers should create their own server implementation or
// use the library directly without HTTP endpoints.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/x1unix/go-cgminer-api/multiminer"
)

func main() {
	var (
		addr       = flag.String("listen", "", "HTTP listen address (overrides config)")
		configFile = flag.String("config", "multiminer.json", "Configuration file path")
		saveConfig = flag.Bool("save-config", false, "Save current configuration to file and exit")
	)
	flag.Parse()

	// Load configuration
	config, err := multiminer.LoadConfigWithEnv(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override listen address from command line if provided
	if *addr != "" {
		config.Server.ListenAddress = *addr
	}

	// Save configuration if requested
	if *saveConfig {
		if err := config.SaveConfig(*configFile); err != nil {
			log.Fatalf("Failed to save configuration: %v", err)
		}
		fmt.Printf("Configuration saved to %s\n", *configFile)
		return
	}

	// Set up logging
	logger := multiminer.NewSimpleLogger(config.GetLogLevel())
	multiminer.SetLogger(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	multiminer.LogInfo(ctx, "Starting multi-miner server",
		multiminer.F("version", "1.0.0"),
		multiminer.F("listen_address", config.Server.ListenAddress),
		multiminer.F("log_level", config.Logging.Level))

	// Create registry and register all drivers
	reg := multiminer.NewRegistry()
	drivers := []multiminer.Driver{
		multiminer.NewCGMinerDriver(),
		multiminer.NewAntminerDriver(),
		multiminer.NewBraiinsDriver(),
		multiminer.NewLuxOSDriver(),
		multiminer.NewWhatsminerDriver(),
		multiminer.NewGoldshellDriver(),
		multiminer.NewHiveOSDriver(),
		multiminer.NewIPolloDriver(),
	}

	for _, driver := range drivers {
		reg.Register(driver)
		multiminer.LogDebug(ctx, "Registered driver", multiminer.F("driver", driver.Name()))
	}

	// Create manager with configuration
	mgr := multiminer.NewManagerWithOptions(reg, config.ToManagerOptions())

	// Start background cleanup if enabled
	if config.Manager.AutoCleanup {
		mgr.StartCleanup(ctx, config.Manager.CleanupInterval)
		multiminer.LogInfo(ctx, "Started background cleanup",
			multiminer.F("interval", config.Manager.CleanupInterval))
	}

	// Ensure cleanup on shutdown
	defer func() {
		multiminer.LogInfo(ctx, "Shutting down manager")
		if err := mgr.Close(); err != nil {
			multiminer.LogError(ctx, "Error closing manager", multiminer.F("error", err))
		}
	}()

	// Load initial devices from environment
	if env := os.Getenv("DEVICES"); env != "" {
		multiminer.LogInfo(ctx, "Loading initial devices from environment")
		for _, item := range strings.Split(env, ",") {
			parts := strings.SplitN(item, "=", 2)
			if len(parts) != 2 {
				continue
			}
			id := multiminer.MinerID(parts[0])
			endpoint := multiminer.Endpoint{Address: parts[1]}

			if err := mgr.AddOrDetect(ctx, id, endpoint, nil); err != nil {
				multiminer.LogWarn(ctx, "Failed to add device from environment",
					multiminer.F("id", string(id)),
					multiminer.F("address", parts[1]),
					multiminer.F("error", err))
			} else {
				multiminer.LogInfo(ctx, "Added device from environment",
					multiminer.F("id", string(id)),
					multiminer.F("address", parts[1]))
			}
		}
	}

	// Create and configure HTTP server
	srv := multiminer.NewServer(mgr)

	httpServer := &http.Server{
		Addr:         config.Server.ListenAddress,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Start server in background
	go func() {
		multiminer.LogInfo(ctx, "Server starting", multiminer.F("address", config.Server.ListenAddress))
		if err := srv.Start(ctx, config.Server.ListenAddress); err != nil && err != http.ErrServerClosed {
			multiminer.LogError(ctx, "Server error", multiminer.F("error", err))
			stop() // Signal shutdown
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	multiminer.LogInfo(ctx, "Shutdown signal received, gracefully shutting down...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		multiminer.LogError(ctx, "Server shutdown error", multiminer.F("error", err))
	}

	multiminer.LogInfo(ctx, "Server stopped")
}
