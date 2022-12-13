// Sample video_quickstart uses the Google Cloud Video Intelligence API to label a video.
package main

import (
	"github.com/payashi/vtrack"
)

var bucketName = "gcs-video-tracking"
var objName = "2022-12-07-0300-1t.json"

func main() {
	vtrack.Extract(bucketName, objName)
}
