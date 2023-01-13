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
	ar1.Plot(outDir, objName1)
	ar2.Plot(outDir, objName2)

	plots := vtrack.NewSyncedPlots(ar1.At(0), ar2.At(10))

	model := vtrack.NewModel(
		vtrack.Config{Phi1: 0., Phi2: 0., K1: 1.28, K2: 0.512, L: 18.97},
	)

	model.Tune(
		plots,
		mat.NewVecDense(4, []float64{ // params
			-0.01 * math.Pi, -0.01 * math.Pi, // theta1, theta2
			2.3, 2.2, // z1, z2
		}),
		1e-2, 1e-2, 100000,
	)
	model.Plot(outDir, "before", plots)
	model.Plot(outDir, "after", plots)
	model.PrintParams(true)
	model.PrintParams(false)
	points := model.Convert(plots)
	for _, p := range points {
		fmt.Printf("%v\n", p)
	}
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
		ar := vtrack.Extract(bucketName, objName)

		// Save on local
		newFile, err := json.MarshalIndent(ar, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = ioutil.WriteFile(filePath, newFile, 0644)
		return ar
	}
}
