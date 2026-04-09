package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/sri/learngo/week24/01_capstone/internal/api"
	"github.com/sri/learngo/week24/01_capstone/internal/model"
	"github.com/sri/learngo/week24/01_capstone/internal/observability"
	"github.com/sri/learngo/week24/01_capstone/internal/service"
	"github.com/sri/learngo/week24/01_capstone/internal/store"
)

// ========================================
// Server Command
// ========================================
// The serve command starts both the REST API and gRPC servers
// with graceful shutdown support. This demonstrates:
//   - Week 6-7: Goroutines, context propagation, signal handling
//   - Week 8-9: HTTP server configuration
//   - Week 15-16: gRPC server setup
//   - Week 17-18: Observability initialization
//   - Week 23: Cobra CLI integration

var (
	httpPort int
	grpcPort int
)

// rootCmd is the base command for the TaskFlow CLI.
var rootCmd = &cobra.Command{
	Use:   "taskflow",
	Short: "TaskFlow — A full-stack Go application",
	Long: `TaskFlow is the capstone project for the 24-week LearnGo curriculum.

It demonstrates production-quality Go patterns including REST APIs,
gRPC services, PostgreSQL and ArangoDB storage, OpenTelemetry tracing,
Prometheus metrics, structured logging, and Docker deployment.

Commands:
  serve    Start the REST and gRPC servers
  migrate  Run database migrations`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST and gRPC servers",
	Long: `Start both the HTTP REST API and gRPC servers.

The servers run concurrently and shut down gracefully on SIGINT/SIGTERM.

Examples:
  taskflow serve
  taskflow serve --http-port 8080 --grpc-port 9090`,

	RunE: runServer,
}

// Execute is the CLI entry point called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// ========================================
	// Configuration Setup
	// ========================================
	cobra.OnInitialize(initConfig)

	// Persistent flags available to all commands
	rootCmd.PersistentFlags().String("config", "", "config file (default: ./config.yaml)")

	// Server-specific flags
	serveCmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP server port")
	serveCmd.Flags().IntVar(&grpcPort, "grpc-port", 9090, "gRPC server port")

	// Bind flags to Viper so config file can override them
	viper.BindPFlag("server.http_port", serveCmd.Flags().Lookup("http-port"))
	viper.BindPFlag("server.grpc_port", serveCmd.Flags().Lookup("grpc-port"))

	// Register commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
}

// initConfig reads the configuration file and environment variables.
func initConfig() {
	cfgFile, _ := rootCmd.Flags().GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./01_capstone")
	}

	// Set defaults for all configuration values
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.grpc_port", 9090)
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.user", "taskflow")
	viper.SetDefault("database.postgres.password", "taskflow")
	viper.SetDefault("database.postgres.dbname", "taskflow")
	viper.SetDefault("database.postgres.sslmode", "disable")
	viper.SetDefault("database.postgres.max_conns", 10)

	viper.SetDefault("database.arango.endpoints", []string{"http://localhost:8529"})
	viper.SetDefault("database.arango.database", "taskflow")
	viper.SetDefault("database.arango.user", "root")
	viper.SetDefault("database.arango.password", "")

	viper.SetDefault("observability.tracing.enabled", true)
	viper.SetDefault("observability.tracing.endpoint", "localhost:4318")
	viper.SetDefault("observability.tracing.service_name", "taskflow")
	viper.SetDefault("observability.metrics.enabled", true)
	viper.SetDefault("observability.logging.level", "info")
	viper.SetDefault("observability.logging.format", "json")

	// Environment variable support (TASKFLOW_SERVER_HTTP_PORT, etc.)
	viper.SetEnvPrefix("TASKFLOW")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		}
	}
}

// ========================================
// Server Startup with Graceful Shutdown
// ========================================
// This function demonstrates several key Go patterns:
//   - Context propagation for cancellation (Week 7)
//   - Goroutines for concurrent server startup (Week 6)
//   - Signal handling for graceful shutdown (Week 7)
//   - WaitGroup for coordinated shutdown (Week 6)
//   - Dependency injection via function parameters

func runServer(cmd *cobra.Command, args []string) error {
	// ========================================
	// 1. Initialize Observability
	// ========================================
	// Observability must be initialized first so all other components
	// can use tracing, metrics, and structured logging from the start.

	logger := observability.NewLogger(
		viper.GetString("observability.logging.level"),
		viper.GetString("observability.logging.format"),
	)
	slog.SetDefault(logger)

	slog.Info("starting TaskFlow",
		"http_port", viper.GetInt("server.http_port"),
		"grpc_port", viper.GetInt("server.grpc_port"),
	)

	// Initialize OpenTelemetry tracing
	ctx := context.Background()
	if viper.GetBool("observability.tracing.enabled") {
		shutdown, err := observability.InitTracer(ctx,
			viper.GetString("observability.tracing.service_name"),
			viper.GetString("observability.tracing.endpoint"),
		)
		if err != nil {
			slog.Warn("tracing initialization failed, continuing without tracing",
				"error", err)
		} else {
			defer shutdown(ctx)
			slog.Info("OpenTelemetry tracing initialized")
		}
	}

	// Initialize Prometheus metrics
	if viper.GetBool("observability.metrics.enabled") {
		observability.RegisterMetrics()
		slog.Info("Prometheus metrics registered")
	}

	// ========================================
	// 2. Initialize Data Stores
	// ========================================
	// We initialize both PostgreSQL (relational) and ArangoDB (graph)
	// connections. In a real app, these would connect to actual databases.
	// For this demo, we log the connection attempt.

	pgStore, err := store.NewPostgresStore(ctx, store.PostgresConfig{
		Host:     viper.GetString("database.postgres.host"),
		Port:     viper.GetInt("database.postgres.port"),
		User:     viper.GetString("database.postgres.user"),
		Password: viper.GetString("database.postgres.password"),
		DBName:   viper.GetString("database.postgres.dbname"),
		SSLMode:  viper.GetString("database.postgres.sslmode"),
		MaxConns: viper.GetInt("database.postgres.max_conns"),
	})
	if err != nil {
		slog.Warn("PostgreSQL connection failed, using in-memory fallback",
			"error", err)
		// In a real app, you might exit here. For the demo, we continue
		// with a nil store and handle it in the service layer.
	} else {
		defer pgStore.Close()
		slog.Info("PostgreSQL connection established")
	}

	arangoStore, err := store.NewArangoStore(ctx, store.ArangoConfig{
		Endpoints: viper.GetStringSlice("database.arango.endpoints"),
		Database:  viper.GetString("database.arango.database"),
		User:      viper.GetString("database.arango.user"),
		Password:  viper.GetString("database.arango.password"),
	})
	if err != nil {
		slog.Warn("ArangoDB connection failed, graph features disabled",
			"error", err)
	} else {
		slog.Info("ArangoDB connection established")
	}

	// ========================================
	// 3. Initialize Service Layer
	// ========================================
	// The service layer contains business logic and orchestrates
	// between multiple stores. This is the dependency injection pattern
	// from Week 13-14.

	taskService := service.NewTaskService(pgStore, arangoStore, logger)
	userService := service.NewUserService(pgStore, logger)

	_ = userService // Used by API handlers below

	// ========================================
	// 4. Create a Root Context with Cancellation
	// ========================================
	// This context is the parent of all server operations.
	// When we cancel it, everything shuts down cleanly.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// ========================================
	// 5. Set Up Signal Handling
	// ========================================
	// Listen for SIGINT (Ctrl+C) and SIGTERM (Docker stop).
	// When received, we initiate graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// ========================================
	// 6. Start HTTP Server
	// ========================================
	httpAddr := fmt.Sprintf(":%d", viper.GetInt("server.http_port"))

	// Build the REST API handler with middleware chain
	restHandler := api.NewRESTHandler(taskService, userService, logger)
	mux := http.NewServeMux()

	// Register REST API routes (Go 1.22+ enhanced routing)
	restHandler.RegisterRoutes(mux)

	// Register Prometheus metrics endpoint
	mux.Handle("GET /metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","version":"%s"}`, model.Version)
	})

	// Apply middleware stack
	handler := api.ChainMiddleware(mux,
		api.RequestIDMiddleware,
		api.LoggingMiddleware(logger),
		api.CORSMiddleware(viper.GetStringSlice("server.cors_origins")),
		api.RecoveryMiddleware(logger),
	)

	readTimeout, _ := time.ParseDuration(viper.GetString("server.read_timeout"))
	writeTimeout, _ := time.ParseDuration(viper.GetString("server.write_timeout"))

	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// ========================================
	// 7. Start gRPC Server
	// ========================================
	grpcAddr := fmt.Sprintf(":%d", viper.GetInt("server.grpc_port"))

	grpcServer := grpc.NewServer()
	api.RegisterGRPCServices(grpcServer, taskService, logger)

	// ========================================
	// 8. Launch Servers Concurrently
	// ========================================
	// Each server runs in its own goroutine. We use a WaitGroup
	// to wait for both to finish during shutdown.
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("HTTP server starting", "addr", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			cancel() // Signal other goroutines to shut down
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			slog.Error("gRPC listener failed", "error", err)
			cancel()
			return
		}
		slog.Info("gRPC server starting", "addr", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
			cancel()
		}
	}()

	// ========================================
	// 9. Wait for Shutdown Signal
	// ========================================
	slog.Info("TaskFlow is running",
		"http", httpAddr,
		"grpc", grpcAddr,
		"pid", os.Getpid(),
	)
	fmt.Printf("\nTaskFlow is running:\n")
	fmt.Printf("  HTTP API: http://localhost%s\n", httpAddr)
	fmt.Printf("  gRPC:     localhost%s\n", grpcAddr)
	fmt.Printf("  Metrics:  http://localhost%s/metrics\n", httpAddr)
	fmt.Printf("  Health:   http://localhost%s/health\n", httpAddr)
	fmt.Printf("\nPress Ctrl+C to stop.\n")

	select {
	case sig := <-sigChan:
		slog.Info("shutdown signal received", "signal", sig)
	case <-ctx.Done():
		slog.Info("context cancelled")
	}

	// ========================================
	// 10. Graceful Shutdown
	// ========================================
	// Give in-flight requests time to complete before forcing shutdown.
	// This is critical for production services — it prevents dropping
	// requests during deployments.

	shutdownTimeout, _ := time.ParseDuration(viper.GetString("server.shutdown_timeout"))
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	slog.Info("initiating graceful shutdown", "timeout", shutdownTimeout)

	// Shut down HTTP server (stops accepting new connections, waits for in-flight)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	} else {
		slog.Info("HTTP server stopped gracefully")
	}

	// Shut down gRPC server (graceful stop waits for RPCs to complete)
	grpcServer.GracefulStop()
	slog.Info("gRPC server stopped gracefully")

	// Cancel the root context to clean up any remaining goroutines
	cancel()

	// Wait for all server goroutines to finish
	wg.Wait()

	slog.Info("TaskFlow shutdown complete")
	return nil
}
