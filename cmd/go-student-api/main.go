package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sandeep-singh-99/go-student-api/internal/config"
	"github.com/Sandeep-singh-99/go-student-api/internal/http/handlers/student"
	"github.com/Sandeep-singh-99/go-student-api/internal/storage/sqlite"
)

func main() {
	// load the configuration
	config := config.MustLoad()

	storage, err := sqlite.New(config)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", config.Env), slog.String("version", "1.0.0"))

	// setup router
	router := http.NewServeMux()

	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))

	// setup server

	server := http.Server{
		Addr:    config.Address,
		Handler: router,
	}

	fmt.Println("Server is running on", config.Address)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-done

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancel()

	// err := server.Shutdown(ctx)

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down server", "error", err)
	} else {
		slog.Info("Server shut down gracefully")
	}
}
