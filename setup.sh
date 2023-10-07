#!/bin/bash

curl -OL https://golang.org/dl/go1.16.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.16.7.linux-amd64.tar.gz

echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export GOROOT=/usr/local/go" >> ~/.bashrc
echo "export PATH=$PATH:$GOPATH/bin:$GOROOT/bin" >> ~/.bashrc
source ~/.bashrc

python3 -m pip install -r apptier/requirements.txt  --no-cache-dir