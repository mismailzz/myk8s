package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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

func CreatePod(w http.ResponseWriter, r *http.Request) {

	// Read request response
	var newPod Pod
	err := json.NewDecoder(r.Body).Decode(&newPod)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Initialize Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Container Image Pull
	imageName := newPod.Image //imageName := "bfirsh/reticulate-splines"

	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Create Container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// Start Container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)

	// Store in pod registry
	newPod.ID = resp.ID
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pods)
}

// SetupRoutes initializes API routes
func SetupRoutes() {
	http.HandleFunc("/createPod", CreatePod)
	http.HandleFunc("/listPods", ListPods)
}

func main() {
	fmt.Println("Starting API Server on port 8080...")
	SetupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
