#!/bin/bash

sudo apt update -y
sudo apt install python3-pip -y
python3 -m pip install -r requirements.txt  --no-cache-dir