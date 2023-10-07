package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type ResponseQueueBody struct {
	Hash   string `json:"hash"`
	Output string `json:"output"`
}

type RequestQueueBody struct {
	Name         string `json:"name"`
	EncodedImage string `json:"encoded_image"`
	Hash         string `json:"hash"`
}

type SQSApi interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)

	ReceiveMessage(ctx context.Context,
		params *sqs.ReceiveMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context,
		params *sqs.DeleteMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

func GetQueueURL(c context.Context, api SQSApi, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return api.GetQueueUrl(c, input)
}

func SendMsg(c context.Context, api SQSApi, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessage(c, input)
}

func GetMessages(c context.Context, api SQSApi, input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return api.ReceiveMessage(c, input)
}
func RemoveMessage(c context.Context, api SQSApi, input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return api.DeleteMessage(c, input)
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
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	fmt.Println("Server is up")
	return
}

func convertImage(file multipart.File) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// Encode as base64.
	contentType := http.DetectContentType(data)

	switch contentType {
	case "image/png":
		fmt.Println("Image type is already PNG.")
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("unable to decode jpeg: %w", err)
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return "", fmt.Errorf("unable to encode png: %w", err)
		}
		data = buf.Bytes()
	default:
		return "", fmt.Errorf("unsupported content typo: %s", contentType)
	}
	imgBase64Str := base64.StdEncoding.EncodeToString(data)
	return imgBase64Str, nil
}

func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func uploadImage(w http.ResponseWriter, r *http.Request, client *sqs.Client) {
	file, hdr, err := r.FormFile("myfile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		log.Fatal(err)
		return
	}
	defer file.Close()

	base64image, err := convertImage(file)
	imageHash := md5.Sum([]byte(base64image))

	requestQueue := "request_queue.fifo"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &requestQueue,
	}

	result, err := GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		fmt.Println("Got an error getting the request queue URL:")
		fmt.Println(err)
		return
	}
	requestqueueURL := result.QueueUrl
	messageBody, _ := json.Marshal(
		RequestQueueBody{hdr.Filename, base64image, hex.EncodeToString(imageHash[:])},
	)
	id := RandStringBytes(6)
	sMInput := &sqs.SendMessageInput{
		MessageBody:    aws.String(string(messageBody)),
		MessageGroupId: &id,
		QueueUrl:       requestqueueURL,
	}

	resp, err := SendMsg(context.TODO(), client, sMInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		fmt.Println("Got an error sending the message:")
		fmt.Println(err)
		return
	}

	fmt.Println("Sent message with ID: " + *resp.MessageId)

	responseQueue := "response_queue.fifo"
	gQInput = &sqs.GetQueueUrlInput{
		QueueName: &responseQueue,
	}

	result, err = GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "" + err.Error()
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		fmt.Println("Got an error getting the response queue URL:")
		fmt.Println(err)
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
		msgResult, _ := GetMessages(context.TODO(), client, gMInput)

		if msgResult.Messages != nil {
			for _, message := range msgResult.Messages {
				var responseBody ResponseQueueBody
				err = json.Unmarshal([]byte(*message.Body), &responseBody)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if responseBody.Hash == hex.EncodeToString(imageHash[:]) {
					dMInput := &sqs.DeleteMessageInput{
						QueueUrl:      responseQueueURL,
						ReceiptHandle: message.ReceiptHandle,
					}

					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/json")
					resp := make(map[string]string)
					resp["message"] = ""
					jsonResp, _ := json.Marshal(resp)
					w.Write(jsonResp)
					_, err = RemoveMessage(context.TODO(), client, dMInput)
					fmt.Println("Deleted message from queue with URL " + *responseQueueURL)
					return
				}
			}
		}

	}
}
