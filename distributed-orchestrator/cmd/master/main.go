package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"distributed-orchestrator/pkg/api"
	"distributed-orchestrator/pkg/scheduler"
	"distributed-orchestrator/pkg/store"
	"distributed-orchestrator/pkg/types"
	
	"github.com/gorilla/mux"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting Distributed Orchestrator Master")
	
	// Initialize etcd store
	etcdEndpoints := []string{"localhost:2379"}
	if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
		etcdEndpoints = []string{endpoints}
	}
	
	etcdStore, err := store.NewEtcdStore(etcdEndpoints, "orchestrator")
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	defer etcdStore.Close()
	
	// Initialize scheduler with bin packing algorithm
	schedulingAlgorithm := scheduler.NewBinPackingAlgorithm()
	taskScheduler := scheduler.NewScheduler(etcdStore, schedulingAlgorithm)
	
	// Start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go taskScheduler.Run(ctx)
	
	// Initialize API server
	apiServer := api.NewServer(etcdStore, taskScheduler)
	
	// Start HTTP server
	router := mux.NewRouter()
	apiServer.RegisterRoutes(router)
	
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	
	// Start gRPC server (for worker communication)
	grpcServer := grpc.NewServer()
	// Register gRPC services here
	
	grpcListener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}
	
	go func() {
		log.Println("Starting gRPC server on :9090")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down...")
	
	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
	cancel()
	
	log.Println("Shutdown complete")
}