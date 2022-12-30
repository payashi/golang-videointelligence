package main

import (
	"github.com/payashi/vtrack"
)

var bucketName = "gcs-video-tracking"

var objName1 = "2022-12-07-0300-1t"
var objName2 = "2022-12-07-0300-2t"

func main() {
	// vtrack.DetectPerson(bucketName, objName)
	trjs1 := vtrack.Extract(bucketName, objName1, true)[0]
	trjs2 := vtrack.Extract(bucketName, objName2, true)[10]
	vtrack.Regress(trjs1, trjs2)
	// vtrack.Demo()
}
