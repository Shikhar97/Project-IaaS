import time
import boto3
import os

AWS_ACCESS_KEY_ID = os.getenv('AWS_ACCESS_KEY')
AWS_SECRET_ACCESS_KEY = os.getenv('AWS_SECRET_ACCESS_KEY')
AWS_REGION = 'us-east-1'  

queue_url = 'https://sqs.us-east-1.amazonaws.com/../RequestQueue.fifo'

MIN_INSTANCES = 1
MAX_INSTANCES = 20

ec2 = boto3.client('ec2', region_name=region_name,
                   aws_access_key_id=aws_access_key_id,
                   aws_secret_access_key=aws_secret_access_key)
sqs = boto3.client('sqs', region_name=region_name,
                   aws_access_key_id=aws_access_key_id,
                   aws_secret_access_key=aws_secret_access_key)

image_id = 'ami-09c6ef0459a2ff40e'
instance_type = 't2.micro'
key_name = 'cc_project1'

def create_instance():
    try:
        instances = ec2.run_instances(
            ImageId=image_id,
            MinCount=1,
            MaxCount=1,
            InstanceType=instance_type,
            KeyName=key_name
        )
        instance_id = instances['Instances'][0]['InstanceId']
        return instance_id
    except Exception as e:
        print(f"Failed to create instance: {e}")
        return None
        
def get_approx_total_msgs():
    try:
        response = sqs.get_queue_attributes(
            QueueUrl=queue_url,
            AttributeNames=['ApproximateNumberOfMessages']
        )
        total_msgs = int(response['Attributes']['ApproximateNumberOfMessages'])
        return total_msgs
    except Exception as e:
        print(f"Failed to get queue attributes: {e}")
        return None
        
def get_num_of_instances():
    try:
        instances = ec2.describe_instance_status(IncludeAllInstances=True)
        running_instances = [i for i in instances['InstanceStatuses'] if
                             i['InstanceState']['Name'] in ('pending', 'running')]
        total_running_instances = len(running_instances)
        return total_running_instances
    except Exception as e:
        print(f"Failed to get instance status: {e}")
        return None

def main():
    while True:
        total_msgs = get_approx_total_msgs()
        total_running_instances = get_num_of_instances()
        total_app_instances = total_running_instances - 1 

        print(f"Messages in Input Queue: {total_msgs}")
        print(f"Total App Instances: {total_app_instances}")

        if total_msgs > 0 and total_msgs > total_app_instances:
            instances_to_launch = min(19 - total_app_instances, total_msgs - total_app_instances)
            if instances_to_launch > 0:
                for i in range(instances_to_launch):
                    instance_id = create_instance()
                    if instance_id:
                        print(f"Launched instance {instance_id}")

        time.sleep(3000) 

if __name__ == "__main__":
    main()