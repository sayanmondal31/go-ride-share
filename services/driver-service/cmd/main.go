package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/driver-service/internal/events"
	grpchandler "ride-sharing/services/driver-service/internal/infrastructure/grpc_handler"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"

	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr = ":9092"
)

func main() {
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
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

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}

	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	service := service.NewService()

	// starting the grpc server
	grpcserver := grpcserver.NewServer()

	grpchandler.NewGrpcHandler(grpcserver, service)

	consumer := events.NewTripConsumer(rabbitmq, service)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("Failed to listen to the message: %v", err)
		}

	}()

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
