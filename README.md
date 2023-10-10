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
1. Input Bucket: inputimage-bucket
2. Output Bucket: outputresults-bucket

## Project Description:
This project is an image classification app that utilizes AWS services and a web tier to perform image classification tasks. Below is a brief overview of the project components:

### Web Tier (Go)

#### setup.sh
Bash script to set up the environment and install necessary packages specified in a ‘requirements.txt’ to run the Go backend.

#### main.go 
This code serves as the backend for an image recognition service. It accepts image uploads, processes them, and communicates asynchronously with AWS SQS queues to manage recognition requests and responses.

### App Tier (Python):

#### setup.sh
Bash installing the Go programming language and configuring the environment variables necessary for Go development.

#### apptier.py
This script listens to an SQS request queue for image classification tasks. When a message arrives in the request queue, it decodes the image, performs classification, stores the image in an S3 bucket, and sends the classification result to a response queue.

#### Image_classification.py
Contains the provided image classification code. 

#### imagenet-labels.json
json file which contains the image labels. 

### Auto Scaling (Python):

#### setup.sh
Bash script to set up the environment and install packages specified in a ‘requirements.txt’.

#### autoscale.py
A module to handle the scale-in and scale-out functionality. 

## How to Run the Project:
1. Clone this repository to your local machine.
2. Set up your AWS credentials (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY).
3. Run the setup scripts to install dependencies for both Python and Go.
4. Deploy the web tier on your server and ensure it's accessible via the provided URL.
5. Use the workload generator to send image classification requests to the web tier.
