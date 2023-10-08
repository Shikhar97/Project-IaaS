import os
import time
import json
import uuid

import boto3
import base64

from dotenv import dotenv_values

from image_classification import classify_image


class SQSApi:

    def __init__(self, conf) -> None:
        # Initializing AWS SQS client
        self.sqs = boto3.client(
            'sqs',
            region_name=conf.get('AWS_DEFAULT_REGION'),
            aws_access_key_id=conf.get('AWS_ACCESS_KEY_ID'),
            aws_secret_access_key=conf.get('AWS_SECRET_ACCESS_KEY')
        )
        print('Initialized SQS client')

    # Function to send messages to response queue
    def sendMessage(self, queue_url, message_body):
        response = self.sqs.send_message(
            QueueUrl=queue_url,
            MessageBody=json.dumps(message_body),
            MessageGroupId=str(uuid.uuid1())
        )

        return response

    # Function to receive messages from request queue
    def receiveMessage(self, queue_url):
        response = self.sqs.receive_message(
            QueueUrl=queue_url,
            MaxNumberOfMessages=1,
            MessageAttributeNames=[
                'All'
            ]
        )

        return response

    # Function to delete messages after it reads from request queue
    def deleteMessage(self, queue_url, receipt_handle):
        self.sqs.delete_message(
            QueueUrl=queue_url,
            ReceiptHandle=receipt_handle
        )


if __name__ == "__main__":
    config = dotenv_values()

    run = SQSApi(config)
    # Initializing S3 bucket client
    s3 = boto3.client(
        's3',
        region_name=config.get('AWS_DEFAULT_REGION'),
        aws_access_key_id=config.get('AWS_ACCESS_KEY_ID'),
        aws_secret_access_key=config.get('AWS_SECRET_ACCESS_KEY')
    )

    print('Initialized S3 client')

    request_queue_url = config.get('REQUEST_QUEUE_URL')
    response_queue_url = config.get('RESPONSE_QUEUE_URL')
    input_bucket_name = config.get('INPUT_BUCKET')
    output_bucket_name = config.get('OUTPUT_BUCKET')

    while True:
        try:
            request_queue_response = run.receiveMessage(request_queue_url)

            # If there is any message in the request queue
            if 'Messages' in request_queue_response and len(request_queue_response['Messages']) > 0:
                print('Message in the request queue found')
                message_body = json.loads(request_queue_response['Messages'][0]['Body'])
                filename = message_body["name"]
                encoded_image = message_body["encoded_image"]
                image_hash = message_body["hash"]

                # Saving the decoded file to temp file
                with open("/tmp/%s" % filename, "wb") as fh:
                    fh.write(base64.b64decode(encoded_image))

                # call image classification function
                value, result = classify_image("/tmp/%s" % filename)

                # uploading file name
                s3.upload_file("/tmp/%s" % filename, input_bucket_name, filename)

                # send message to response queue
                response = run.sendMessage(
                    response_queue_url,
                    {
                        "Hash": image_hash,
                        "Output": result
                    })
                # uploading file name and result
                s3.put_object(Bucket=output_bucket_name, Key=filename.split(".")[0], Body=value)

                # delete the message from the request queue
                run.deleteMessage(request_queue_url, request_queue_response['Messages'][0]['ReceiptHandle'])
            else:
                print('No message in the request queue.')
                # Polling every 60 seconds to check for messages
                time.sleep(60)
        except Exception as e:
            print('Something went wrong with request queue response receive message.')
            continue
