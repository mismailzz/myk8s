package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Pod represents a simple containerized application
type Pod struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// Store running pods in memory (for simplicity)
var podRegistry = make(map[string]Pod)

var dockerClient *client.Client // Global Docker client

// Initialize Docker client
func init() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to initialize Docker client: %s", err)
	}
}

// CreatePod starts a new container using Docker
func CreatePod(w http.ResponseWriter, r *http.Request) {
	var newPod Pod

	// Read request body
	err := json.NewDecoder(r.Body).Decode(&newPod)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Schedule the pod using the scheduler
	err = SchedulePod(&newPod)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to schedule pod: %s", err), http.StatusInternalServerError)
		return
	}

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

	// Respond with pod list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pods)
}

// DeletePod stops and removes a container
func DeletePod(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing pod ID", http.StatusBadRequest)
		return
	}

	// Stop the container
	if err := dockerClient.ContainerStop(context.Background(), id, container.StopOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop container '%s': %s", id, err), http.StatusInternalServerError)
		return
	}

	// Remove the container (SDK requires an empty struct)
	if err := dockerClient.ContainerRemove(context.Background(), id, container.RemoveOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove container '%s': %s", id, err), http.StatusInternalServerError)
		return
	}

	// Remove from registry
	delete(podRegistry, id)

	// Log success
	log.Printf("Pod deleted: ID=%s", id)

	// Respond to client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pod %s deleted", id)
}

// SetupRoutes initializes API routes
func SetupRoutes() {
	http.HandleFunc("/createPod", CreatePod)
	http.HandleFunc("/listPods", ListPods)
	http.HandleFunc("/deletePod", DeletePod)
}

func main() {
	fmt.Println("Starting API Server on port 8080...")
	SetupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
