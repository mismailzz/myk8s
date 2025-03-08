package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Pod represents a simple containerized application
type Pod struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// Store running pods in memory (for simplicity)
var podRegistry = make(map[string]Pod)

// CreatePod handles pod creation requests
func CreatePod(w http.ResponseWriter, r *http.Request) {
	var newPod Pod
	err := json.NewDecoder(r.Body).Decode(&newPod) // Read JSON body
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Assign a unique ID (for now, using the pod name as ID)
	newPod.ID = newPod.Name
	podRegistry[newPod.ID] = newPod

	// Respond to client
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPod)
}

// ListPods returns all running pods
func ListPods(w http.ResponseWriter, r *http.Request) {
	pods := make([]Pod, 0, len(podRegistry))
	for _, pod := range podRegistry {
		pods = append(pods, pod)
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pods)
}

// DeletePod handles pod deletion
func DeletePod(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id") // Read pod ID from query param
	if id == "" {
		http.Error(w, "Missing pod ID", http.StatusBadRequest)
		return
	}

	_, exists := podRegistry[id]
	if !exists {
		http.Error(w, "Pod not found", http.StatusNotFound)
		return
	}

	delete(podRegistry, id) // Remove from registry
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pod %s deleted", id)
}

// SetupRoutes initializes API routes
func SetupRoutes() {
	http.HandleFunc("/createPod", CreatePod)
	http.HandleFunc("/listPods", ListPods)
	http.HandleFunc("/deletePod", DeletePod)
}

// Main function to start API server
func main() {
	fmt.Println("Starting API Server on port 8080...")
	SetupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil)) // Start HTTP server
}
