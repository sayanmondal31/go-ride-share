package main

import (
	"encoding/json"
	"log"
	"net/http"
	grpcclients "ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {

	var reqBody previewTripRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	// close after unmershal done
	defer r.Body.Close()

	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	// jsonBody, _ := json.Marshal(reqBody)
	// reader := bytes.NewReader(jsonBody)

	tripService, err := grpcclients.NewTripServiceClient()

	if err != nil {
		log.Fatal(err)
	}

	defer tripService.Close()

	// tripService.Client.PreviewTrip()

	// resp, err := http.Post("http://trip-service:8083/preview", "application/json", reader)

	// if err != nil {
	// 	log.Print(err)
	// 	return
	// }

	// defer resp.Body.Close()

	tripPreview, err := tripService.Client.PreviewTrip(r.Context(), reqBody.ToProto())

	if err != nil {
		log.Printf("Failed to preview a trip: %v", err)
		http.Error(w, "Failed to preview trip", http.StatusInternalServerError)
		return
	}

	// var respBody any

	// if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
	// 	http.Error(w, "failed to parse JSON data from trip service", http.StatusBadRequest)
	// 	return
	// }

	response := contracts.APIResponse{
		Data: tripPreview,
	}

	writeJSON(w, http.StatusCreated, response)
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed tp parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if reqBody.UserId == "" && reqBody.RideFareID == "" {
		http.Error(w, "userId and ridefaireId required!", http.StatusBadRequest)
		return
	}

	tripService, err := grpcclients.NewTripServiceClient()

	if err != nil {
		log.Fatal()
	}

	defer tripService.Close()

	tripStart, err := tripService.Client.CreateTrip(r.Context(), reqBody.ToProto())

	if err != nil {
		log.Printf("Failed to preview a trip: %v", err)
	}

	response := contracts.APIResponse{
		Data: tripStart,
	}

	writeJSON(w, http.StatusCreated, response)

}
