// Sample video_quickstart uses the Google Cloud Video Intelligence API to label a video.
package main

import (
	"github.com/payashi/vtrack"
)

var bucketName = "gcs-video-tracking"

var objName1 = "2022-12-07-0300-1t"
var objName2 = "2022-12-07-0300-2t"

func main() {
	// vtrack.DetectPerson(bucketName, objName)
	trjs1 := vtrack.Extract(bucketName, objName1, true)
	trjs2 := vtrack.Extract(bucketName, objName2, true)
	vtrack.Regress(trjs1[0], trjs2[10])
	// vtrack.Demo()
}
