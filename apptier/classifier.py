import boto3
import os
import time
from dotenv import load_dotenv
from image_classification_modified import ClassifyImage


load_dotenv()

AWS_ACCESS_KEY = os.getenv('AWS_ACCESS_KEY')
AWS_SECRET_ACCESS_KEY = os.getenv('AWS_SECRET_ACCESS_KEY')

class TestSQS():

    def __init__(self) -> None:
        pass

    def sendMessage(self, queue_url, sqs, image_name, result):
        response = sqs.send_message(
            QueueUrl=queue_url,
            MessageAttributes={
                'Title': {
                    'DataType': 'String',
                    'StringValue': image_name
                }
            },
            MessageBody=result,
            MessageGroupId='images'
        )

        return response

    def receiveMessage(self, queue_url, sqs):
        response = sqs.receive_message(
            QueueUrl=queue_url,
            AttributeNames=[
                'SentTimestamp'
            ],
            MaxNumberOfMessages=1,
            MessageAttributeNames=[
                'All'
            ]
        )
    
        return response

    def deleteMessage(self, queue_url, sqs, receipt_handle):
        sqs.delete_message(
            QueueUrl=queue_url,
            ReceiptHandle=receipt_handle
        )

    def putBucket(self, s3, bucket_name, image_name, result):
        response = s3.put_object(Bucket=bucket_name, Key=image_name, Body=result)
        return response

if __name__ == "__main__":
    run = TestSQS()
    sqs = boto3.client(
        'sqs',
         region_name='us-east-1',
         aws_access_key_id=AWS_ACCESS_KEY,
         aws_secret_access_key=AWS_SECRET_ACCESS_KEY
	)
    s3 = boto3.client(
        's3',
         region_name='us-east-1',
         aws_access_key_id=AWS_ACCESS_KEY,
         aws_secret_access_key=AWS_SECRET_ACCESS_KEY
    )

    print('Initialized boto3 clients.')

    request_queue_url = 'https://sqs.us-east-1.amazonaws.com/../RequestQueue.fifo'
    response_queue_url = 'https://sqs.us-east-1.amazonaws.com/../ResponseQueue.fifo'
    get_bucket_name = 'input-images-cc-2023'
    put_bucket_name = 'output-bucket-cc-2023'

    # receive message from the queue
    while(True): 
        print('Running next iteration.')
        time.sleep(5)
        request_queue_response = {}

        try:
            request_queue_response = run.receiveMessage(request_queue_url, sqs)
            print('request queue response', request_queue_response)
        except:
            print('Something went wrong with request queue response receive message.')
            continue

        # do image classification
        if 'Messages' in request_queue_response and len(request_queue_response['Messages']) > 0:
            # gray_image = ClassifyImage.getImage(get_bucket_name, request_queue_response['Messages'][0]['Body'])
            result = ClassifyImage.classifyImage(get_bucket_name, request_queue_response['Messages'][0]['Body'])
            print('result', result)

            image_name = request_queue_response['Messages'][0]['Body'].rstrip('.JPEG')
            print('image_name', image_name)

            # send message to response queue
            response = run.sendMessage(response_queue_url, sqs, request_queue_response['Messages'][0]['Body'], result)
            print(response['MessageId'])
            print('Sent message to response queue.')

            # put the object to the output bucket
            repsonse = run.putBucket(s3, put_bucket_name, image_name, result)

            print('Uploaded image to output S3 bucket.')

            #delete the message from the request queue
            run.deleteMessage(request_queue_url, sqs, request_queue_response['Messages'][0]['ReceiptHandle']) 
            print('Deleted message from request queue.')
        else: 
            print('No message in the request queue.')
    
