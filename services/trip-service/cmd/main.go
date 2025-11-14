package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr = ":9093"
)

func main() {

	inmemRepo := repository.NewInmemRepository()

	svc := service.NewService(inmemRepo)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel() // clean resources

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// starting the grpc server
	grpcserver := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcserver, svc)

	log.Printf("Starting grpc server Trip service on port %s", lis.Addr().String())

	go func() {
		if err := grpcserver.Serve(lis); err != nil {
			log.Fatalf("Failed to server:  %v", err)
			cancel()
		}
	}()

	// wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down the server")
	grpcserver.GracefulStop()

}
