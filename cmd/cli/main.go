package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	
	"distributed-orchestrator/pkg/store"
	"distributed-orchestrator/pkg/types"
	
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "orchestrator-cli",
		Usage:    "Command line interface for the orchestrator",
		Version:  "1.0.0",
		Compiled: time.Now(),
		Commands: []*cli.Command{
			{
				Name:    "task",
				Aliases: []string{"t"},
				Usage:   "Task management",
				Subcommands: []*cli.Command{
					{
						Name:    "create",
						Aliases: []string{"c"},
						Usage:   "Create a new task",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "image",
								Aliases:  []string{"i"},
								Usage:    "Container image",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Task name",
							},
							&cli.StringSliceFlag{
								Name:    "command",
								Aliases: []string{"c"},
								Usage:   "Command to execute",
							},
							&cli.StringSliceFlag{
								Name:    "env",
								Aliases: []string{"e"},
								Usage:   "Environment variables (KEY=VALUE)",
							},
							&cli.Int64Flag{
								Name:  "cpu",
								Usage: "CPU requirement (millicores)",
								Value: 1000,
							},
							&cli.Int64Flag{
								Name:  "memory",
								Usage: "Memory requirement (MB)",
								Value: 512,
							},
						},
						Action: createTask,
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List tasks",
						Action:  listTasks,
					},
					{
						Name:    "get",
						Aliases: []string{"g"},
						Usage:   "Get task details",
						Action:  getTask,
					},
					{
						Name:    "delete",
						Aliases: []string{"d"},
						Usage:   "Delete a task",
						Action:  deleteTask,
					},
				},
			},
			{
				Name:    "node",
				Aliases: []string{"n"},
				Usage:   "Node management",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List nodes",
						Action:  listNodes,
					},
					{
						Name:    "get",
						Aliases: []string{"g"},
						Usage:   "Get node details",
						Action:  getNode,
					},
					{
						Name:   "drain",
						Usage:  "Drain a node",
						Action: drainNode,
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "etcd-endpoints",
				Aliases: []string{"e"},
				Usage:   "ETCD endpoints",
				Value:   "localhost:2379",
			},
		},
	}
	
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getStore(c *cli.Context) (store.Store, error) {
	endpoints := []string{c.String("etcd-endpoints")}
	return store.NewEtcdStore(endpoints, "orchestrator")
}

func createTask(c *cli.Context) error {
	s, err := getStore(c)
	if err != nil {
		return err
	}
	defer s.Close()
	
	env := make(map[string]string)
	for _, envStr := range c.StringSlice("env") {
		// Parse KEY=VALUE format
		// Simplified parsing
	}
	
	task := &types.Task{
		ID:        types.GenerateID(),
		Name:      c.String("name"),
		Image:     c.String("image"),
		Command:   c.StringSlice("command"),
		Env:       env,
		Resources: types.ResourceRequirements{
			CPU:    c.Int64("cpu"),
			Memory: c.Int64("memory"),
		},
		State:        types.TaskStatePending,
		CreationTime: time.Now().Unix(),
	}
	
	if err := s.StoreTask(context.Background(), task); err != nil {
		return err
	}
	
	fmt.Printf("Task created: %s\n", task.ID)
	return nil
}

func listTasks(c *cli.Context) error {
	s, err := getStore(c)
	if err != nil {
		return err
	}
	defer s.Close()
	
	tasks, err := s.GetAllTasks(context.Background())
	if err != nil {
		return err
	}
	
	for _, task := range tasks {
		fmt.Printf("%s: %s (%s)\n", task.ID, task.Name, task.State)
	}
	return nil
}

// Similar implementations for getTask, deleteTask, listNodes, getNode, drainNode