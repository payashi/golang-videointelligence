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
	ar1 := getAnnotationResults(objName1)
	ar2 := getAnnotationResults(objName2)
	model := vtrack.NewModel(ar1.At(0), ar2.At(10))
	var z0 float64 = -0.1
	model.Plot2D(z0, outDir, "before")
	model.BatchGradientDecent(1e-2, 1e-1, 100000)
	model.Plot2D(z0, outDir, "after")
	model.PrintParams()
}

func getAnnotationResults(objName string) vtrack.AnnotationResults {
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
