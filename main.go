package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// --- CONFIGURATION ---
const (
	WorkerCount = 3               // Simulating a node that can only handle 3 tasks at once
	MemoryLimit = 128 * 1024 * 1024 // 128MB Hard Limit per task
)

// --- MONITOR LOGIC  ---
type MonitorStats struct {
	MemoryUsageMB float64
}

// GetStats reads the Docker Engine API (which proxies Linux cgroups)
func GetStats(ctx context.Context, cli *client.Client, containerID string) (*MonitorStats, error) {
	statsStream, err := cli.ContainerStats(ctx, containerID, false) // stream=false for snapshot
	if err != nil {
		return nil, err
	}
	defer statsStream.Body.Close()

	var v types.StatsJSON
	if err := json.NewDecoder(statsStream.Body).Decode(&v); err != nil {
		return nil, err
	}

	// Calculate Memory Usage in MB
	memUsage := float64(v.MemoryStats.Usage) / 1024 / 1024
	return &MonitorStats{MemoryUsageMB: memUsage}, nil
}

// --- SCHEDULER LOGIC  ---
type Task struct {
	ID    string
	Image string
	Cmd   string
}

type Scheduler struct {
	DockerCli *client.Client
	Queue     chan Task
	Wg        sync.WaitGroup
}

func (s *Scheduler) RunWorker(workerID int) {
	defer s.Wg.Done()
	ctx := context.Background()

	for task := range s.Queue {
		fmt.Printf("[Worker %d] Starting %s...\n", workerID, task.ID)

		// 1. Create Container with RESOURCE LIMITS
		resp, err := s.DockerCli.ContainerCreate(ctx,
			&container.Config{
				Image: task.Image,
				Cmd:   []string{"sh", "-c", task.Cmd},
			},
			&container.HostConfig{
				Resources: container.Resources{
					Memory: MemoryLimit, // Enforce 128MB limit
				},
			}, nil, nil, task.ID)

		if err != nil {
			log.Printf("[Error] Failed to create %s: %v", task.ID, err)
			continue
		}

		// 2. Start Container
		if err := s.DockerCli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			log.Printf("[Error] Failed to start %s: %v", task.ID, err)
			continue
		}

		// 3. Monitor Resource Usage
		time.Sleep(1 * time.Second) // Give it a second to spin up
		if stats, err := GetStats(ctx, s.DockerCli, resp.ID); err == nil {
			fmt.Printf("[Monitor] %s is using %.2f MB RAM\n", task.ID, stats.MemoryUsageMB)
		}

		// 4. Wait for Completion
		statusCh, errCh := s.DockerCli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				log.Printf("[Error] Wait error on %s: %v", task.ID, err)
			}
		case <-statusCh:
			fmt.Printf("[Worker %d] %s Finished.\n", workerID, task.ID)
		}

		// 5. Cleanup
		s.DockerCli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	}
}

// --- MAIN ENTRY POINT ---
func main() {
	// CLI Arguments
	count := flag.Int("count", 5, "Number of tasks to run")
	image := flag.String("image", "alpine", "Docker image to use")
	cmd := flag.String("cmd", "sleep 2", "Command to run inside container")
	flag.Parse()

	// Connect to Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to connect to Docker: %v", err)
	}

	fmt.Println("--- Mini-Scheduler Initialized ---")
	fmt.Printf("Config: %d Workers | %d Tasks | Image: %s\n", WorkerCount, *count, *image)

	// Initialize Scheduler
	scheduler := &Scheduler{
		DockerCli: cli,
		Queue:     make(chan Task, 100),
	}

	// Start Workers (The "Round Robin" Logic)
	for i := 1; i <= WorkerCount; i++ {
		scheduler.Wg.Add(1)
		go scheduler.RunWorker(i)
	}

	// Submit Jobs
	go func() {
		for i := 1; i <= *count; i++ {
			taskID := fmt.Sprintf("job-%d", i)
			scheduler.Queue <- Task{ID: taskID, Image: *image, Cmd: *cmd}
			fmt.Printf("[API] Submitted %s\n", taskID)
		}
		close(scheduler.Queue)
	}()

	// Wait for everything to finish
	scheduler.Wg.Wait()
	fmt.Println("--- All Jobs Completed ---")
}