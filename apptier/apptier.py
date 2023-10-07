import os
import time
import json
import boto3
import base64

from dotenv import load_dotenv, find_dotenv, dotenv_values


# from image_classification import classify_image


class SQSApi:

    def __init__(self, conf) -> None:
        self.sqs = boto3.client(
            'sqs',
            region_name=conf.get('AWS_DEFAULT_REGION'),
            aws_access_key_id=conf.get('AWS_ACCESS_KEY_ID'),
            aws_secret_access_key=conf.get('AWS_SECRET_ACCESS_KEY')
        )

    def sendMessage(self, queue_url, message_body):
        response = self.sqs.send_message(
            QueueUrl=queue_url,
            MessageBody=json.dumps(message_body),
        )

        return response

    def receiveMessage(self, queue_url):
        response = self.sqs.receive_message(
            QueueUrl=queue_url,
            MaxNumberOfMessages=1,
            MessageAttributeNames=[
                'All'
            ]
        )

        return response

    def deleteMessage(self, queue_url, receipt_handle):
        self.sqs.delete_message(
            QueueUrl=queue_url,
            ReceiptHandle=receipt_handle
        )

    def putBucket(self, s3, bucket_name, image_name, result):
        response = s3.put_object(Bucket=bucket_name, Key=image_name, Body=result)
        return response


if __name__ == "__main__":
    config = dotenv_values()

    run = SQSApi(config)

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

    # receive message from the queue
    # while True:
    #     print('Running next iteration.')
    #     time.sleep(5)
    #     request_queue_response = {}
    #
    #     try:
    #         request_queue_response = run.receiveMessage(request_queue_url)
    #         print('request queue response', request_queue_response)
    #     except Exception as e:
    #         print('Something went wrong with request queue response receive message.')
    #         continue
    #
    #     # do image classification
    #     if 'Messages' in request_queue_response and len(request_queue_response['Messages']) > 0:
    #
    #         message_body = request_queue_response['Messages'][0]['Body']
    #
    #         filename = message_body["Name"]
    #         encoded_image = message_body["EncodedImage"]
    #         image_hash = message_body["Hash"]
    #         print('image_name', filename)
    #
    #         with open("/tmp/%s" % filename, "wb") as fh:
    #             fh.write(base64.b64decode(encoded_image))
    #
    #         result = classify_image("/tmp/%s" % filename)
    #         print('result', result)
    #
    #         s3.upload_file("/tmp/%s" % filename, input_bucket_name, filename)
    #         print('Uploaded image to output S3 bucket.')
    #
    #         # send message to response queue
    #         response = run.sendMessage(
    #             response_queue_url,
    #             {
    #                 "Hash": image_hash,
    #                 "Output": result
    #             })
    #         print(response['MessageId'])
    #         print('Sent message to response queue.')
    #         s3.put_object(Bucket=output_bucket_name, Key=filename.split(".")[0], Body=result)
    #
    #         # delete the message from the request queue
    #         run.deleteMessage(request_queue_url, request_queue_response['Messages'][0]['ReceiptHandle'])
    #         print('Deleted message from request queue.')
    #     else:
    #         print('No message in the request queue.')
