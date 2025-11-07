package main

import (
	"fmt"
	"log"
	"net/http"
	h "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8083")
)

func main() {

	inmemRepo := repository.NewInmemRepository()

	svc := service.NewService(inmemRepo)

	mux := http.NewServeMux()

	// register handlers
	httpHandler := h.HttpHandler{Service: svc}

	// register preview endpoint; method filtering should be done inside the handler
	mux.HandleFunc("/preview", httpHandler.HandleTripPreview)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server error %v", err)
	}

	fmt.Println("hello")
}
