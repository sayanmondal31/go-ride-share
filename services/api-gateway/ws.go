package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"

	grpcclients "ride-sharing/services/api-gateway/grpc_clients"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRiderWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
		return
	}

	userID := r.URL.Query().Get("userID") // this what frontend will send

	if userID == "" {
		log.Println("No user ID provided")
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message:  %v", err)
			break
		}

		log.Printf("Received message: %s", message)

	}

	defer conn.Close()

}

func handleDriverWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")

	if packageSlug == "" {
		log.Println("No package slug provided")
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID") // this what frontend will send

	if userID == "" {
		log.Println("No user ID provided")
		return
	}

	// type Driver struct {
	// 	Id             string `json:"id"`
	// 	Name           string `json:"name"`
	// 	ProfilePicture string `json:"profilePicture"`
	// 	CarPlate       string `json:"carPlate"`
	// 	PackageSlug    string `json:"packageSlug"`
	// }
	ctx := r.Context()
	driverService, err := grpcclients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	defer driverService.Close()
	defer func() {
		driverService.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		driverService.Close()
		log.Println("Driver unregistered: ", userID)

	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})

	if err != nil {
		log.Printf("Error registering driver : %v", err)
		return
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message:  %v", err)
			break
		}

		log.Printf("Received message: %s", message)

	}

}
