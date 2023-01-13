package vtrack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"sort"

	"cloud.google.com/go/storage"
	"cloud.google.com/go/videointelligence/apiv1/videointelligencepb"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const AspectRatio float64 = 16. / 9.
const MaxDur int = 601 // time range

func Extract(bucketName string, objName string) AnnotationResults {
	// Download a json file
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bkt := client.Bucket(bucketName)

	// Get the bucket's attributes
	attrs, err := bkt.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bucket %s, created at %s, is located in %s with storage class %s\n",
		attrs.Name, attrs.Created, attrs.Location, attrs.StorageClass)

	// Unmarshal a json file
	obj := bkt.Object(objName + ".json")
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	slurp, _ := io.ReadAll(r)
	var res videointelligencepb.AnnotateVideoResponse
	json.Unmarshal(slurp, &res)
	if err := json.Unmarshal(slurp, &res); err != nil {
		panic(err)
	}

	// Translate AnnotateVideoResponse to Trajectory object
	annots := res.AnnotationResults[0].PersonDetectionAnnotations
	var ret AnnotationResults
	ret.trajectories = make([]Trajectory, len(annots))
	for i, annot := range annots {
		track := annot.Tracks[0]
		tj := &ret.trajectories[i]
		tj.plots = make([][]float64, MaxDur)
		tj.conf = track.Confidence
		tj.start = track.Segment.StartTimeOffset.AsDuration().Milliseconds() / 100
		tj.end = track.Segment.EndTimeOffset.AsDuration().Milliseconds() / 100

		for _, tsobj := range track.TimestampedObjects {
			box := tsobj.NormalizedBoundingBox
			tidx := tsobj.TimeOffset.AsDuration().Milliseconds() / 100
			tj.plots[tidx] = []float64{
				float64((box.Left+box.Right)/2) - 0.5,
				(0.5 - float64(box.Top)) / AspectRatio,
			}
		}
		tj.length = tj.calcLength()
	}
	sort.Slice(ret.trajectories, func(i, j int) bool {
		return ret.trajectories[i].length > ret.trajectories[j].length
	})
	return ret
}

type Trajectory struct {
	conf          float32
	plots         [][]float64
	start, end    int64
	length        float64
	width, height float64
}

type AnnotationResults struct {
	trajectories []Trajectory
}

func (tj Trajectory) Trimmedplots() [][]float64 {
	return tj.plots[tj.start : tj.end+1]

}

func (tj Trajectory) calcLength() float64 {
	ret := .0
	for i := int(tj.start); i < int(tj.end); i++ {
		cx, cy := tj.plots[i][0], tj.plots[i][1]
		nx, ny := tj.plots[i+1][0], tj.plots[i+1][1]
		dist := math.Sqrt((nx-cx)*(nx-cx) + (ny-cy)*(ny-cy))
		ret += dist
	}
	return ret
}

func (ar AnnotationResults) Plot(outDir, fileName string) {
	p := plot.New()

	// p.Title.Text = "Trajectories"
	// p.X.Label.Text = "X"
	// p.Y.Label.Text = "Y"

	p.X.Min = -0.5
	p.Y.Min = -0.5 / AspectRatio
	p.X.Max = +0.5
	p.Y.Max = +0.5 / AspectRatio

	p.Add(plotter.NewGrid())

	for i, tj := range ar.trajectories[:20] {
		plots := make(plotter.XYs, len(tj.Trimmedplots()))
		for i, v := range tj.Trimmedplots() {
			plots[i].X = v[0]
			plots[i].Y = v[1]
		}
		ploti, err := plotter.NewScatter(plots)
		if err != nil {
			panic(err)
		}

		ploti.GlyphStyle.Color = plotutil.Color(i)
		ploti.GlyphStyle.Radius = 2
		p.Add(ploti)

		p.Legend.Add(fmt.Sprintf("tr-%2d [%03d-%03d]", i, tj.start, tj.end), ploti)
	}
	pwidth := 6 * vg.Inch
	pheight, _ := vg.ParseLength(fmt.Sprintf("%.2fin", 6/AspectRatio))

	if err := p.Save(pwidth, pheight, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (tj Trajectory) MarshalJSON() ([]byte, error) {
	v := &struct {
		Conf   float32     `json:"conf"`
		Plots  [][]float64 `json:"plots"`
		Start  int64       `json:"start"`
		End    int64       `json:"end"`
		Length float64     `json:"length"`
		Width  float64     `json:"width"`
		Height float64     `json:"height"`
	}{
		Conf:   tj.conf,
		Plots:  tj.plots,
		Start:  tj.start,
		End:    tj.end,
		Length: tj.length,
		Width:  tj.width,
		Height: tj.height,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (tj *Trajectory) UnmarshalJSON(b []byte) error {
	tj2 := &struct {
		Conf   float32     `json:"conf"`
		Plots  [][]float64 `json:"plots"`
		Start  int64       `json:"start"`
		End    int64       `json:"end"`
		Length float64     `json:"length"`
		Width  float64     `json:"width"`
		Height float64     `json:"height"`
	}{}
	err := json.Unmarshal(b, tj2)
	tj.conf = tj2.Conf
	tj.plots = tj2.Plots
	tj.start = tj2.Start
	tj.end = tj2.End
	tj.length = tj2.Length
	tj.width = tj2.Width
	tj.height = tj2.Height
	return err
}

func (ar AnnotationResults) At(index int) Trajectory {
	return ar.trajectories[index]
}

func (ar AnnotationResults) MarshalJSON() ([]byte, error) {
	v := &struct {
		Trajectories []Trajectory `json:"trajectories"`
	}{ar.trajectories}
	s, err := json.Marshal(v)
	return s, err
}

func (ar *AnnotationResults) UnmarshalJSON(b []byte) error {
	ar2 := &struct {
		Trajectories []Trajectory `json:"trajectories"`
	}{}
	err := json.Unmarshal(b, ar2)
	ar.trajectories = ar2.Trajectories
	return err
}
