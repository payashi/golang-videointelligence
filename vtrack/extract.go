package vtrack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	"cloud.google.com/go/storage"
	"cloud.google.com/go/videointelligence/apiv1/videointelligencepb"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var maxdur int32 = 700

const width float64 = 1280
const height float64 = 720

func Extract(bucketName string, objName string, plot bool) []Trajectory {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bkt := client.Bucket(bucketName)
	attrs, err := bkt.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bucket %s, created at %s, is located in %s with storage class %s\n",
		attrs.Name, attrs.Created, attrs.Location, attrs.StorageClass)
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
	annots := res.AnnotationResults[0].PersonDetectionAnnotations
	trjs := make([]Trajectory, len(annots))
	for i, annot := range annots {
		track := annot.Tracks[0]
		tj := &trjs[i]
		tj.Plots = make(plotter.XYs, maxdur)
		tj.Conf = track.Confidence
		tj.Start = track.Segment.StartTimeOffset.AsDuration().Milliseconds() / 100
		tj.End = track.Segment.EndTimeOffset.AsDuration().Milliseconds() / 100

		for _, tsobj := range track.TimestampedObjects {
			box := tsobj.NormalizedBoundingBox
			tidx := tsobj.TimeOffset.AsDuration().Milliseconds() / 100
			tj.Plots[tidx].X = float64((box.Left+box.Right)/2) * width
			tj.Plots[tidx].Y = (1 - float64(box.Top)) * height
		}
		tj.Length = GetLength(*tj)
		tj.Width = width
		tj.Height = height
	}
	sort.Slice(trjs, func(i, j int) bool { return trjs[i].Length > trjs[j].Length })
	if plot {
		Plot(trjs, objName)
	}
	return trjs
}

type Trajectory struct {
	Conf         float32
	Plots        plotter.XYs
	Start, End   int64
	Length       float64
	Width, Height float64
}

func (tr Trajectory) TrimmedPlots() plotter.XYs {
	return tr.Plots[tr.Start : tr.End+1]

}

func GetLength(tr Trajectory) float64 {
	ret := .0
	for i := int(tr.Start); i < int(tr.End); i++ {
		cx, cy := tr.Plots.XY(i)
		nx, ny := tr.Plots.XY(i + 1)
		dist := math.Sqrt((nx-cx)*(nx-cx) + (ny-cy)*(ny-cy))
		ret += dist
	}
	return ret
}

func Plot(trjs []Trajectory, fileName string) {
	rand.Seed(time.Now().UnixNano())
	p := plot.New()

	// p.Title.Text = "Trajectories"
	// p.X.Label.Text = "X"
	// p.Y.Label.Text = "Y"

	p.X.Min = 0
	p.Y.Min = 0
	p.X.Max = width
	p.Y.Max = height

	p.Add(plotter.NewGrid())

	for i, tr := range trjs[:20] {
		ploti, err := plotter.NewScatter(tr.TrimmedPlots())
		if err != nil {
			panic(err)
		}

		ploti.GlyphStyle.Color = plotutil.Color(i)
		ploti.GlyphStyle.Radius = 2
		p.Add(ploti)

		p.Legend.Add(fmt.Sprintf("tr-%2d [%03d-%03d]", i, tr.Start, tr.End), ploti)
	}
	pwidth := 6 * vg.Inch
	pheight, _ := vg.ParseLength(fmt.Sprintf("%.2fin", height/width*6))

	if err := p.Save(pwidth, pheight, fmt.Sprintf("%s.png", fileName)); err != nil {
		panic(err)
	}
}
