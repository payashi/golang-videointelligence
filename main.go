package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/payashi/vtrack"
)

var outDir = "/workspace/golang-videointelligence/out"

var bucketName = "gcs-video-tracking"

var objName1 = "2022-12-07-0300-1t"
var objName2 = "2022-12-07-0300-2t"

func main() {
	tj1 := getTrajectories(objName1)
	tj2 := getTrajectories(objName2)
	tj1.Temp()
	tj2.Temp()
	// vtrack.Demo()
}

func getTrajectories(objName string) vtrack.AnnotationResults {
	filePath := fmt.Sprintf("%s/%s.json", outDir, objName)
	if file, err := os.Open(filePath); err == nil {
		fmt.Printf("File exists\n")
		defer file.Close()
		b, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		var ret vtrack.AnnotationResults
		if err := json.Unmarshal(b, &ret); err != nil {
			panic(err)
		}
		return ret
	} else {
		fmt.Printf("Fetching %s...\n", objName)
		ar := vtrack.Extract(bucketName, objName, false)

		// Save on local
		newFile, err := json.MarshalIndent(ar, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = ioutil.WriteFile(filePath, newFile, 0644)
		return ar
	}
}
