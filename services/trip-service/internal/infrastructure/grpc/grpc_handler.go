package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"

	pb "ride-sharing/shared/proto/trip"

	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer

	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {

	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	pickupCoords := &types.Coordinate{
		Latitude:  pickup.Latitude,
		Longitude: pickup.Longitude,
	}

	destCoord := &types.Coordinate{
		Latitude:  destination.Latitude,
		Longitude: destination.Longitude,
	}

	route, err := h.service.GetRoute(ctx, pickupCoords, destCoord)

	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	userId := req.GetUserId()

	// 1. Estimate the ride fares prices based on the route (ex: distance)
	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)

	// 2. Store the ride fares for the the create the trip to fetch and variables
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userId, route)

	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to generate ride fare: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareId()
	userID := req.GetUserId()

	// 1. Fetch and validate the fare
	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)

	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to  validate fare: %v", err)
	}

	trip, err := h.service.CreateTrip(ctx, rideFare)

	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
		// Trip:   trip.ToProto(),
	}, nil

	// 2. Call creare trip

	// 3. we also need to initialize an empty driver to the trip
	// 4. Add a comment at the end of the function to publish an event on the async comms module

	// return nil, status.Errorf(codes.Unimplemented, "method CreateTrip not implemented")
}
