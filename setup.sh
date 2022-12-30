#!/bin/bash

GOOGLE_APPLICATION_CREDENTIALS="/workspace/golang-videointelligence/cred.json"
echo "export GOOGLE_APPLICATION_CREDENTIALS=\"$GOOGLE_APPLICATION_CREDENTIALS\"" >> ~/.bashrc
echo "export GOPATH=/workspace" >> ~/.bashrc
exec $SHELL -l
# gcloud init --project=payashi-playground
mkdir -p /workspace/golang-videointelligence/out