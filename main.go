package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	model := getModel("model", ar1, ar2)
	model.PrintParams(false)
	tdplots := model.Idenitfy(ar1, ar2)
	vtrack.Plot(outDir, "3dplots", tdplots)
}

func getModel(objName string, ar1, ar2 vtrack.AnnotationResults) *vtrack.Model {
	filePath := fmt.Sprintf("%s/%s.json", outDir, objName)
	if file, err := os.Open(filePath); err == nil {
		defer file.Close()
		b, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		var ret vtrack.Model
		if err := json.Unmarshal(b, &ret); err != nil {
			panic(err)
		}
		return &ret
	} else {
		fmt.Printf("Fetching %s...\n", objName)
		plots, _ := vtrack.NewSyncedPlots(ar1.At(0), ar2.At(10))

		z0 := 1.7
		ret := vtrack.NewModel(
			vtrack.Config{
				K1: 1.32, K2: 0.467,
				C1: *mat.NewVecDense(3, []float64{
					0, 0, 4.028,
				}),
				C2: *mat.NewVecDense(3, []float64{
					0, -18.97, 3.904,
				}),
			},
		)

		ret.Plot(outDir, "before", z0, plots)
		ret.Tune(100000, 1e-2, 1e-2, z0, plots)
		ret.Plot(outDir, "after", z0, plots)

		// Save on local
		newFile, err := json.MarshalIndent(ret, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = ioutil.WriteFile(filePath, newFile, 0644)
		return ret
	}
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
		ar.Plot(outDir, objName)

		// Save on local
		newFile, err := json.MarshalIndent(ar, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = ioutil.WriteFile(filePath, newFile, 0644)
		return ar
	}
}
