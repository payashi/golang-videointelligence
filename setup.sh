#!/bin/bash

GOOGLE_APPLICATION_CREDENTIALS="/workspace/video-tracking/cred.json"
echo "export GOOGLE_APPLICATION_CREDENTIALS=\"$GOOGLE_APPLICATION_CREDENTIALS\"" >> ~/.bashrc
exec $SHELL -l
gcloud init --project=payashi-playground