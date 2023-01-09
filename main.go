package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/payashi/vtrack"
	"gonum.org/v1/gonum/mat"
)

var outDir = "/workspace/golang-videointelligence/out"
var bucketName = "gcs-video-tracking"
var objName1 = "2022-12-07-0300-1t"
var objName2 = "2022-12-07-0300-2t"

func main() {
	ar1 := getAnnotationResults(objName1)
	ar2 := getAnnotationResults(objName2)
	// for _, z0 := range LinSpace(-0.2, -1.2, 21) {
	z0 := -0.5
	params := mat.NewVecDense(6, []float64{
		-0.01 * math.Pi, -0.01 * math.Pi, // theta
		-0.25 * math.Pi, 0.23 * math.Pi, // phi
		1,  // k
		z0, // z0
	})
	model := vtrack.NewModel(ar1.At(0), ar2.At(10), params)
	model.Plot2D(outDir, "before")
	model.BatchGradientDecent(1e-2, 1e-1, 50000)
	model.Plot2D(outDir, "after")
	model.PrintParams(true)
	model.PrintParams(false)
}

func LinSpace(start, end float64, num int) []float64 {
	ret := make([]float64, num)
	for i := 0; i < num; i++ {
		ret[i] = start + (end-start)*float64(i)/float64(num-1)
	}
	return ret
}

func getAnnotationResults(objName string) vtrack.AnnotationResults {
	filePath := fmt.Sprintf("%s/%s.json", outDir, objName)
	if file, err := os.Open(filePath); err == nil {
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
