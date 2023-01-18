package main

import (
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
	// Ntrials: 0,
	Dp: 1e-2,
	Mu: 1e-2,
	Z0: 1.7,
}

func main() {
	// vannotate.DetectPerson(bucketName, objName1)
	srList1 := vannotate.GetSeries(outDir, bucketName, objName1)
	srList2 := vannotate.GetSeries(outDir, bucketName, objName2)

	// TODO: Implement LoadCameraSystem(outDir, fileName) and file existence check should be done on main
	cs := vtrack.GetCameraSystem(outDir, "camsys", srList1[4], srList2[19], config, tconfig)
	cs.PrintUnityParams()
	cs.Plot(outDir, "tdplots", srList1, srList2)
	// vtrack.Plot(outDir, "tdplots", tdplots)
	// ipList := cs.Idenitfy(srList1, srList2)
	// cs.Plot(outDir, "model", )

	// Save on local
	// newFile, err := json.MarshalIndent(ipList, "", "\t")
	// if err != nil {
	// 	panic(err)
	// }
	// err = ioutil.WriteFile(fmt.Sprintf("%s/tdplots.json", outDir), newFile, 0644)
	// if err != nil {
	// 	panic(err)
	// }
}
