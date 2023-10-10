# CSE546-Project1-IaaS

## Project Description
This project is completed in a group of three students. It consists of three-tier architecute (web-tier,
app-tier and the data-storage) along with an auto-scaler to scale in and scale out with respect to the load.

### WebTier
The webtier is a GoLang HTTP server which provides an end point for the user to upload the images. These images are then encoded using base64 and sent to a request SQS queue.

### AppTier
The app-tier keeps polling(long-polling) the request queue and once it receives an image, it classifies the image using the 
given deep learning model. It then returns the results to the response queue and as well as upload to s3 input bucket. 
It also uploads the output from the model to a different bucket.

### AutoScaler
It keeps polling the size of the request queue once the size reaches a threshold it scales out and similarly once the threshold is less it scales in.

### DataStorage
The data storage tier consists of two S3 buckets, one to store input by user and other to store the images_name and the class.

# Getting Started
1. For webtier and autoscaler :
       
        1. Clone this repository to your machine. 
        2. Copy the AWS credentials to a .env file inside webtier/ and autoscaler/ directories.
        3. Run the ./setup.sh scripts to install dependencies for both Python and Go.
        4. To build a binary from the main.go, we can simply run `go build .` when we are inside the webtier directory. This will create a binary file named ./webtier.
        5. After this, we can simply execute the binary by running ./webtier. This will start up the Go server at port 8001.
        6. User can then send requests at http://<public-ip>:8001/upload_images
        7. Run python3 autoscale.py to start the autoscaler script.

2. For app-tier :

        1. Clone this repository to your machine. 
        2. Copy the AWS credentials to a .env file inside webtier/ and autoscaler/ directories.
        3. Run the ./setup.sh scripts to install dependencies for Python.
        3. After this, we can simply start the script using python2 apptier.py.
        3. We added a crontab entry for the script to start on boot:
                     @reboot /usr/bin/python3 /home/ubuntu/Project-IaaS/apptier/apptier.py 
