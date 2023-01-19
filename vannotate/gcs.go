package vannotate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
	video "cloud.google.com/go/videointelligence/apiv1"
	"cloud.google.com/go/videointelligence/apiv1/videointelligencepb"
	videopb "cloud.google.com/go/videointelligence/apiv1/videointelligencepb"
)

func SaveToGCS(bucketName string, objName string) {
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

func loadFromGCS(bucketName string, objName string) []Series {
	const maxDur = 601
	// Download a json file
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bkt := client.Bucket(bucketName)

	// Get the bucket's attributes
	attrs, err := bkt.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bucket %s, created at %s, is located in %s with storage class %s\n",
		attrs.Name, attrs.Created, attrs.Location, attrs.StorageClass)

	// Unmarshal a json file
	obj := bkt.Object(objName + ".json")
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	slurp, _ := io.ReadAll(r)
	var res videointelligencepb.AnnotateVideoResponse
	json.Unmarshal(slurp, &res)
	if err := json.Unmarshal(slurp, &res); err != nil {
		panic(err)
	}

	// Translate AnnotateVideoResponse to Series object
	annots := res.AnnotationResults[0].PersonDetectionAnnotations
	ret := make([]Series, len(annots))
	for i, annot := range annots {
		track := annot.Tracks[0]
		tj := &ret[i]
		tj.Plots = make([]ScreenPlot, maxDur)
		tj.Conf = track.Confidence
		tj.Start = int(track.Segment.StartTimeOffset.AsDuration().Milliseconds() / 100)
		tj.End = int(track.Segment.EndTimeOffset.AsDuration().Milliseconds() / 100)

		for _, tsobj := range track.TimestampedObjects {
			box := tsobj.NormalizedBoundingBox
			tidx := tsobj.TimeOffset.AsDuration().Milliseconds() / 100
			tj.Plots[tidx] = ScreenPlot{
				float64((box.Left+box.Right)/2) - 0.5,
				0.5 - float64(box.Top),
			}
		}
	}
	return ret
}

func GetSeries(outDir, bucketName, objName string) []Series {
	filePath := fmt.Sprintf("%s/%s.json", outDir, objName)
	// If file is found, read and unmarshal it
	if file, err := os.Open(filePath); err == nil {
		defer file.Close()
		b, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		var ret []Series
		if err := json.Unmarshal(b, &ret); err != nil {
			panic(err)
		}
		return ret
	} else {
		fmt.Printf("Fetching %s...\n", objName)
		series := loadFromGCS(bucketName, objName)
		PlotScreen(outDir, objName, series)

		// Save on local
		newFile, err := json.MarshalIndent(series, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = ioutil.WriteFile(filePath, newFile, 0644)
		return series
	}
}
