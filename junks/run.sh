#! /bin/bash

TOKEN=$(gcloud auth application-default print-access-token)

OPERATION_NAME=$(curl -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json; charset=utf-8" \
    -d @request.json \
    "https://videointelligence.googleapis.com/v1/videos:annotate" \
    | jq -r '.name')

curl -X GET \
  -H "Authorization: Bearer "$TOKEN \
  https://videointelligence.googleapis.com/v1/$OPERATION_NAME \
  > response.json
