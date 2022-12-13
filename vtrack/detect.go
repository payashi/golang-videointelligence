package vtrack

import (
	"context"
	"fmt"
	"log"

	video "cloud.google.com/go/videointelligence/apiv1"
	videopb "cloud.google.com/go/videointelligence/apiv1/videointelligencepb"
)

func DetectPerson(bucketName string, objName string) {
	var inputUri string = fmt.Sprintf("gs://%s/%s.mp4", bucketName, objName)
	var outputUri string = fmt.Sprintf("gs://%s/%s.json", bucketName, objName)
	ctx := context.Background()

	// Creates a client.
	client, err := video.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	op, err := client.AnnotateVideo(ctx, &videopb.AnnotateVideoRequest{
		InputUri:  inputUri,
		OutputUri: outputUri,
		Features: []videopb.Feature{
			videopb.Feature_PERSON_DETECTION,
		},
		VideoContext: &videopb.VideoContext{
			PersonDetectionConfig: &videopb.PersonDetectionConfig{
				IncludeAttributes:    true,
				IncludeBoundingBoxes: true,
				IncludePoseLandmarks: true,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to start annotation job: %v", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to annotate: %v", err)
	}
	ndetects := len(resp.AnnotationResults[0].PersonDetectionAnnotations)
	fmt.Printf("%d detections", ndetects)
}
