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
	// track pods to node mapping
	pods := make([]map[string]string, 0, len(podRegistry))
	for _, pod := range podRegistry {
		pods = append(pods, map[string]string{
			"ID":    pod.ID,
			"Name":  pod.Name,
			"Image": pod.Image,
			"Node":  podToNodeMapping[pod.ID],
		})
	}

	// Respond with pod list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pods)
}

// DeletePod stops and removes a container
func DeletePod(w http.ResponseWriter, r *http.Request) {
	// Get pod ID from query string
	podID := r.URL.Query().Get("id")
	if podID == "" {
		http.Error(w, "Missing pod ID", http.StatusBadRequest)
		return
	}

	// Stop the container
	if err := dockerClient.ContainerStop(context.Background(), podID, container.StopOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop container '%s': %s", podID, err), http.StatusInternalServerError)
		return
	}

	// Remove the container (SDK requires an empty struct)
	if err := dockerClient.ContainerRemove(context.Background(), podID, container.RemoveOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove container '%s': %s", podID, err), http.StatusInternalServerError)
		return
	}

	// Update node usage
	// Free up capacity from the node
	if nodeName, exists := podToNodeMapping[podID]; exists {
		for i := range nodes {
			if nodes[i].Name == nodeName {
				nodes[i].Used--
				break
			}
		}
		delete(podToNodeMapping, podID)
	}

	// Remove from registry
	delete(podRegistry, podID)

	// Log success
	log.Printf("Pod deleted: ID=%s", podID)

	// Respond to client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pod %s deleted", podID)
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
