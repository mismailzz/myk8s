package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

// Node
type Node struct {
	Name     string
	Capacity int
	Used     int
}

var node = Node{
	Name:     "Node-1",
	Capacity: 10, // Let's assume each node can handle 10 pods
	Used:     0,
}

// SchedulePod schedules a new pod to a node
func SchedulePod(newPod *Pod) error {

	if node.Used >= node.Capacity {
		return fmt.Errorf("node is full — cannot schedule pod")
	}

	/* SETUP A CONTAINER */ //DockerClient will be intialize in apiserver - no needed it here

	ctx := context.Background()

	// Pull the container image if not available locally
	log.Printf("Pulling image: %s", newPod.Image)
	out, err := dockerClient.ImagePull(ctx, newPod.Image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("Failed to pull image '%s': %s", newPod.Image, err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Create the container
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: newPod.Image,
	}, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("Failed to create container using image '%s': %s", newPod.Image, err)

	}

	// Start the container (SDK requires an empty struct)
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("Failed to start container with ID '%s': %s", resp.ID, err)

	}

	// Store in pod registry
	newPod.ID = resp.ID
	podRegistry[newPod.ID] = *newPod

	log.Printf("Pod %s scheduled to node %s (used: %d/%d)", newPod.Name, node.Name, node.Used, node.Capacity)
	return nil

}
