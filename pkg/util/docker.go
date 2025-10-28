package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	
	"distributed-orchestrator/pkg/types"
	
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

func ExecuteDockerTask(task *types.Task) (string, int, error) {
	ctx := context.Background()
	
	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// Pull image if needed
	_, _, err = cli.ImageInspectWithRaw(ctx, task.Image)
	if err != nil {
		if client.IsErrNotFound(err) {
			// Pull the image
			out, err := cli.ImagePull(ctx, task.Image, types.ImagePullOptions{})
			if err != nil {
				return "", -1, fmt.Errorf("failed to pull image: %v", err)
			}
			defer out.Close()
			io.Copy(io.Discard, out) // Wait for pull to complete
		} else {
			return "", -1, fmt.Errorf("failed to inspect image: %v", err)
		}
	}

	// Create container
	containerConfig := &container.Config{
		Image:      task.Image,
		Cmd:        task.Command,
		Env:        convertEnvMapToSlice(task.Env),
		Labels:     map[string]string{"orchestrator.task.id": task.ID},
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			NanoCPUs: task.Resources.CPU * 1000000, // Convert to nanoseconds
			Memory:   task.Resources.Memory * 1024 * 1024, // Convert to bytes
		},
		AutoRemove:  true,
		NetworkMode: "none", // Isolated network
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", -1, fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", -1, fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for container to complete
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	var statusCode int64
	
	select {
	case err := <-errCh:
		if err != nil {
			return "", -1, fmt.Errorf("error waiting for container: %v", err)
		}
	case result := <-statusCh:
		statusCode = result.StatusCode
	}

	// Get container logs
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", int(statusCode), fmt.Errorf("failed to get container logs: %v", err)
	}
	defer out.Close()

	logs, err := io.ReadAll(out)
	if err != nil {
		return "", int(statusCode), fmt.Errorf("failed to read container logs: %v", err)
	}

	return string(logs), int(statusCode), nil
}

func convertEnvMapToSlice(env map[string]string) []string {
	var envSlice []string
	for key, value := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}
	return envSlice
}

func GetDockerInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	info, err := cli.Info(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"version":       info.ServerVersion,
		"containers":    info.Containers,
		"running":       info.ContainersRunning,
		"paused":        info.ContainersPaused,
		"stopped":       info.ContainersStopped,
		"images":        info.Images,
		"memory":        info.MemTotal,
		"cpus":          info.NCPU,
		"architecture":  info.Architecture,
		"os":            info.OperatingSystem,
	}, nil
}