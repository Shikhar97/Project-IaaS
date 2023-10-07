#!/bin/bash

curl -OL https://go.dev/dl/go1.20.9.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.20.9.linux-amd64.tar.gz

echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export GOROOT=/usr/local/go" >> ~/.bashrc
echo "export PATH=$PATH:$HOME/go/bin:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc