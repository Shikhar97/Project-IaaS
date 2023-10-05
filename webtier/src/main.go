package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type ServerStatus struct {
	Status string `json:"status"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	awsSqsClient := sqs.NewFromConfig(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/", serverCheck)
	mux.HandleFunc("/upload_image", func(writer http.ResponseWriter, request *http.Request) {
		uploadImage(writer, request, awsSqsClient)
	})

	// Create a server instance
	server := &http.Server{
		Addr:    ":8001", // Change the port if needed
		Handler: mux,
	}

	// Handle graceful shutdown on SIGINT and SIGTERM signals
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		fmt.Println("Shutting down gracefully...")
		if err := server.Close(); err != nil {
			fmt.Printf("Error during server shutdown: %v\n", err)
		}
	}()

	fmt.Println("Server listening on :8001")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Error: %v\n", err)
	}
}

func serverCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var serverStatus ServerStatus
	serverStatus.Status = "Server is up"
	err := json.NewEncoder(w).Encode(serverStatus)
	if err != nil {
		return
	}
	fmt.Printf("Server is up\n")
}

func uploadImage(w http.ResponseWriter, r *http.Request, client *sqs.Client) {

}
