# CSE546-Project1-IaaS

## Group Members
1. Shikhar Gupta - Web Tier
2. Maitry Ronakbhai Trivedi - Autoscaling
3. Ayushi Agarwal - App Tier

## AWS Credentials:
- AWS_ACCESS_KEY_ID: []
- AWS_SECRET_ACCESS_KEY: [] 

## PEM Key for Web-Tier SSH Access:
- Key Name: cc_project1.pem

## Web Tier's URL and EIP (Elastic IP):
- Web Tier URL: []
- Elastic IP (EIP): []

## SQS Queue Names:
1. Request Queue: request_queue.fifo
2. Response Queue: response_queue.fifo

## S3 Bucket Names:
1. Input Bucket: []
2. Output Bucket: []

## Project Description:
This project is an image classification app that utilizes AWS services and a web tier to perform image classification tasks. Below is a brief overview of the project components:

### App Tier (Python):
- `apptier.py`: This Python script handles the communication with AWS services. It sends and receives messages from the request and response queues, processes images, and uploads results to S3.

### Image Classification (Python):
- `image_classification.py`: This Python script contains the image classification logic using a pre-trained ResNet-18 model. It classifies the images and returns the result.

### Auto Scaling (Python):
- `autoscale.py`: This Python script implements auto-scaling for the app tier based on the number of messages in the request queue.

### Web Tier (Go):
- `main.go`: This Go script handles incoming HTTP requests, receives images, and sends them to the app tier via AWS SQS. It then waits for and retrieves the classification results from the response queue.

## How to Run the Project:
1. Clone this repository to your local machine.
2. Set up your AWS credentials (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY).
3. Run the setup scripts to install dependencies for both Python and Go.
4. Deploy the web tier on your server and ensure it's accessible via the provided URL.
5. Use the workload generator to send image classification requests to the web tier.
