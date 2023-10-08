package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/joho/godotenv"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// ResponseQueueBody : structure for ResponseQueue
type ResponseQueueBody struct {
	Hash   string `json:"hash"`
	Output string `json:"output"`
}

// RequestQueueBody : structure for RequestQueue
type RequestQueueBody struct {
	Name         string `json:"name"`
	EncodedImage string `json:"encoded_image"`
	Hash         string `json:"hash"`
}

// Load the environment variables from .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
}
func main() {
	// Creating config to create a SQS Client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	// Create a AWS SQS Client
	awsSqsClient := sqs.NewFromConfig(cfg)

	// Creating a http server on port 8001 to handle incoming requests
	mux := http.NewServeMux()
	mux.HandleFunc("/", serverCheck)
	mux.HandleFunc("/upload_image", func(writer http.ResponseWriter, request *http.Request) {
		uploadImage(writer, request, awsSqsClient)
	})

	server := &http.Server{
		Addr:    ":8001",
		Handler: mux,
	}

	// Handle graceful shutdown on SIGINT and SIGTERM signals
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		log.Println("Shutting down gracefully...")
		if err := server.Close(); err != nil {
			log.Printf("Error during server shutdown: %v\n", err)
		}
	}()

	log.Println("Server listening on :8001")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error: %v\n", err)
	}
}

// function to return the status of the server
func serverCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s\n", err)
	}
	w.Write(jsonResp)
	log.Println("Server is up")
	return
}

// function to convert the image to base64 string
func convertImage(file multipart.File) (string, string) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err.Error()
	}

	// Encode as base64.
	contentType := http.DetectContentType(data)

	switch contentType {
	case "image/png":
		log.Println("Image type is already PNG.")
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			errorMsg := "unable to decode jpeg: " + err.Error()
			return "", errorMsg
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			errorMsg := "unable to encode png: " + err.Error()
			return "", errorMsg
		}
		data = buf.Bytes()
	default:
		errorMsg := "unsupported content typo: " + contentType
		return "", errorMsg
	}
	imgBase64Str := base64.StdEncoding.EncodeToString(data)
	return imgBase64Str, ""
}

// function to generate random string for Queue ID
func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// function to send the image to request queue and get a response back
func uploadImage(w http.ResponseWriter, r *http.Request, client *sqs.Client) {
	file, hdr, err := r.FormFile("myfile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["error"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		log.Println(err)
		return
	}
	defer file.Close()

	base64image, msg := convertImage(file)
	if msg != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["error"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		log.Println(err)
		return
	}
	imageHash := md5.Sum([]byte(base64image))

	requestQueue := "request_queue.fifo"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &requestQueue,
	}

	result, err := client.GetQueueUrl(context.TODO(), gQInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["error"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		log.Println("Got an error getting the request queue URL:")
		log.Println(err)
		return
	}
	requestQueueURL := result.QueueUrl
	messageBody, _ := json.Marshal(
		RequestQueueBody{hdr.Filename, base64image, hex.EncodeToString(imageHash[:])},
	)
	id := randStringBytes(6)
	sMInput := &sqs.SendMessageInput{
		MessageBody:    aws.String(string(messageBody)),
		MessageGroupId: &id,
		QueueUrl:       requestQueueURL,
	}

	resp, err := client.SendMessage(context.TODO(), sMInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["error"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		log.Println("Got an error sending the message:")
		log.Println(err)
		return
	}

	log.Println("Sent message with ID: " + *resp.MessageId)

	responseQueue := "response_queue.fifo"
	gQInput = &sqs.GetQueueUrlInput{
		QueueName: &responseQueue,
	}

	result, err = client.GetQueueUrl(context.TODO(), gQInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["error"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		log.Println("Got an error getting the response queue URL:")
		log.Println(err)
		return
	}
	responseQueueURL := result.QueueUrl
	for {
		gMInput := &sqs.ReceiveMessageInput{
			MessageAttributeNames: []string{
				string(types.QueueAttributeNameAll),
			},
			QueueUrl:            responseQueueURL,
			MaxNumberOfMessages: 10,
		}
		msgResult, err := client.ReceiveMessage(context.TODO(), gMInput)
		if err != nil {
			log.Println("Got an error receiving the message:")
			log.Println(err)
		}

		if msgResult.Messages != nil {
			for _, message := range msgResult.Messages {
				var responseBody ResponseQueueBody
				err = json.Unmarshal([]byte(*message.Body), &responseBody)
				if err != nil {
					log.Println(err)
					continue
				}
				if responseBody.Hash == hex.EncodeToString(imageHash[:]) {
					dMInput := &sqs.DeleteMessageInput{
						QueueUrl:      responseQueueURL,
						ReceiptHandle: message.ReceiptHandle,
					}
					_, err := client.DeleteMessage(context.TODO(), dMInput)
					if err != nil {
						log.Println(err)
					}
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "text/plain")
					w.Write([]byte(responseBody.Output))
					return
				}
			}
		}

	}
}
