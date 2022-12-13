package vtrack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

func Extract(bucketName string, objName string) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bkt := client.Bucket(bucketName)
	attrs, err := bkt.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bucket %s, created at %s, is located in %s with storage class %s\n",
		attrs.Name, attrs.Created, attrs.Location, attrs.StorageClass)
	obj := bkt.Object(objName)
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	if _, err := io.Copy(os.Stdout, r); err != nil {
		log.Fatal(err)
	}
	slurp, _ := io.ReadAll(r)

	var res Result

	json.Unmarshal(slurp, &res)
	fmt.Println(res)
}

type Result map[string]interface{}

// // Only one video was processed, so get the first result.
// result := resp.GetAnnotationResults()[0]

// for _, annotation := range result.PersonDetectionAnnotations[3:4] {
// 	for _, track := range annotation.Tracks {
// 		fmt.Printf("\tConfidence: %f\n", track.Confidence)
// 		for _, tsobj := range track.TimestampedObjects {
// 			box := tsobj.NormalizedBoundingBox
// 			t := int32(tsobj.TimeOffset.Seconds) + tsobj.TimeOffset.Nanos/int32(1e8)
// 			fmt.Printf(
// 				"\t\t(%.2f, %.2f) @%d\n",
// 				box.Top, (box.Left+box.Right)/2, t,
// 			)
// 		}
// 	}
// }
