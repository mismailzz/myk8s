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

// Multiple nodes
var nodes = []Node{
	{Name: "node1", Capacity: 2, Used: 0},
	{Name: "node2", Capacity: 2, Used: 0},
	{Name: "node3", Capacity: 2, Used: 0},
}

var nodeIndex = 0 // For round-robin scheduling
// Pod-to-Node mapping
var podToNodeMapping = make(map[string]string)

// SchedulePod schedules a new pod to a node
func SchedulePod(newPod *Pod) error {

	// Select a node
	node, err := selectNode()
	if err != nil {
		return err
	}

	/* SETUP A CONTAINER */ //DockerClient will be intialize in apiserver - no needed it here

	ctx := context.Background()

	// Pull the container image if not available locally
	log.Printf("Pulling image: %s", newPod.Image)
	out, err := dockerClient.ImagePull(ctx, newPod.Image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image '%s': %s", newPod.Image, err)
	}
	defer out.Close()

	io.Copy(os.Stdout, out)

	// Create the container
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: newPod.Image,
	}, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container using image '%s': %s", newPod.Image, err)

	}

	// Start the container (SDK requires an empty struct)
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container with ID '%s': %s", resp.ID, err)

	}

	// Store in pod registry
	newPod.ID = resp.ID
	podRegistry[newPod.ID] = *newPod

	// Update node usage
	node.Used++
	podToNodeMapping[newPod.ID] = node.Name

	log.Printf("Pod %s scheduled to node %s (used: %d/%d)", newPod.Name, node.Name, node.Used, node.Capacity)
	return nil

}

// selectNode selects the least-loaded node
func selectNode() (*Node, error) {

	for i := 0; i < len(nodes); i++ {
		node := &nodes[nodeIndex]
		nodeIndex = (nodeIndex + 1) % len(nodes)
		if node.Used < node.Capacity {
			return node, nil
		}

	}
	return nil, fmt.Errorf("all nodes are full, cannot schedule pod")

}
