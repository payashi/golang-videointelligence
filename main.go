package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/payashi/vannotate"
	"github.com/payashi/vtrack"
	"gonum.org/v1/gonum/mat"
)

var outDir = "/workspace/golang-videointelligence/out"
var bucketName = "gcs-video-tracking"
var objName1 = "2022-12-07-0300-1t"
var objName2 = "2022-12-07-0300-2t"

var config = vtrack.Config{
	K1: 1.32, K2: 0.467,
	R1: 16. / 9., R2: 16. / 9.,
	C1: *mat.NewVecDense(3, []float64{
		0, 0, 4.028,
	}),
	C2: *mat.NewVecDense(3, []float64{
		0, -18.97, 3.904,
	}),
}

var tconfig = vtrack.TuneConfig{
	Ntrials: 100000,
	Dp:      1e-2,
	Mu:      1e-2,
	Z0:      1.7,
}

func main() {
	// vannotate.DetectPerson(bucketName, objName1)
	srList1 := vannotate.GetSeries(outDir, bucketName, objName1)
	srList2 := vannotate.GetSeries(outDir, bucketName, objName2)

	filePath := fmt.Sprintf("%s/%s.json", outDir, "camsys")
	cs, err := vtrack.LoadCameraSystem(filePath)
	if err != nil {
		fmt.Printf("Tuning the Camera System...\n")
		sr1, sr2 := srList1[0], srList2[0]
		plots, _ := vtrack.NewSyncedPlots(sr1, sr2)
		tconfig.Plots = plots

		cs = vtrack.NewCameraSystem(config)
		cs.Tune(tconfig)
		// cs.Plot(fmt.Sprintf("%s/%s", outDir, "after.png"), []vannotate.Series{sr1}, []vannotate.Series{sr2})

		// Save on local
		newFile, err := json.MarshalIndent(cs, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = os.WriteFile(filePath, newFile, 0644)
	}

	cs.PrintUnityParams()
	cs.Plot(fmt.Sprintf("%s/%s.png", outDir, "iplots"), srList1, srList2)

	ipList := cs.Idenitfy(srList1, srList2)
	cs.PlotJoined(fmt.Sprintf("%s/%s.png", outDir, "joined"), ipList[:3])

	// Save on local
	newFile, err := json.MarshalIndent(ipList, "", "\t")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fmt.Sprintf("%s/iplots.json", outDir), newFile, 0644)
	if err != nil {
		panic(err)
	}
}
