package vtrack

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// TODO: draw frames

func NewModel(tj1, tj2 Trajectory, data ...*mat.VecDense) *Model {
	start := MaxInt(tj1.start, tj2.start)
	end := MinInt(tj1.end, tj2.end)
	if start > end {
		log.Fatal("These Trajectories have no common segments.")
	}
	tr1 := tj1.plots[start : end+1]
	tr2 := tj2.plots[start : end+1]
	model := new(Model)
	model.ndata = int(end - start + 1) // the number of datasets
	model.data = make([]Snapshot, model.ndata)
	for i := 0; i < model.ndata; i++ {
		p1, q1 := tr1[i][0], tr1[i][1]
		p2, q2 := tr2[i][0], tr2[i][1]
		model.data[i] = Snapshot{
			p1/tj1.width - 0.5,
			p2/tj2.width - 0.5,
			(q1/tj1.height - 0.5) * (tj1.height / tj1.width),
			(q2/tj2.height - 0.5) * (tj2.height / tj2.width),
		}
	}
	model.c1 = mat.NewVecDense(3, []float64{0, 0, 0})
	model.c2 = mat.NewVecDense(3, []float64{0, 1, 0})
	if len(data) > 0 {
		model.params = data[0]
	} else {
		model.params = mat.NewVecDense(5, []float64{
			-0.1, -0.1, // theta
			0.7 * math.Pi, 1.2 * math.Pi,
			10, // k
		})
	}
	model.nparams = model.params.Len()
	return model
}

type Snapshot struct {
	p1, p2, q1, q2 float64
}

type Model struct {
	nparams, ndata int
	params         *mat.VecDense
	data           []Snapshot
	c1, c2         *mat.VecDense
}

func (model Model) PrintParams() {
	fmt.Printf("theta1: %.2f pi, theta2: %.2f pi\t",
		model.params.At(0, 0)/math.Pi,
		model.params.At(1, 0)/math.Pi,
	)
	fmt.Printf("phi1: %.2f pi, phi2: %.2f pi\t",
		model.params.At(2, 0)/math.Pi,
		model.params.At(3, 0)/math.Pi,
	)
	fmt.Printf("k: %.2f\n",
		model.params.At(4, 0),
	)
}

func (model *Model) NaiveGradientDecent(dp, mu float64, ntrials int) {
	for i := 0; i < ntrials; i++ {
		model.PrintParams()
		for j := 0; j < model.nparams; j++ {
			for k := 0; k < 100; k++ {
				inc := mat.NewVecDense(model.nparams, nil)
				inc.SetVec(j, -mu*model.GetDiff(j, dp))
				model.params.AddVec(model.params, inc)
			}
		}
	}
}

func (model *Model) BatchGradientDecent(dp, mu float64, ntrials int) {
	for i := 0; i < ntrials; i++ {
		inc := mat.NewVecDense(model.nparams, nil)
		for j := 0; j < model.nparams; j++ {
			inc.SetVec(j, -model.GetDiff(j, dp))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		model.params.AddScaledVec(model.params, mu*math.Exp(-4*float64(i)/float64(ntrials)), inc)
	}
}

func (model *Model) GetDiff(i int, dp float64) float64 {
	cv := model.GetPointsDistance(-0.1, model.params)
	nparams := mat.NewVecDense(model.nparams, nil)
	nparams.SetVec(i, dp)
	nparams.AddVec(model.params, nparams)
	nv := model.GetPointsDistance(-0.1, nparams)
	return (nv - cv) / dp
}

func (model *Model) project(z0 float64, params *mat.VecDense) [][]float64 {
	theta1, theta2 := params.At(0, 0), params.At(1, 0)
	phi1, phi2 := params.At(2, 0), params.At(3, 0)
	k := params.At(4, 0)

	n1 := mat.NewVecDense(3, []float64{
		math.Cos(phi1) * math.Cos(theta1),
		math.Sin(phi1) * math.Cos(theta1),
		math.Sin(theta1),
	})
	n2 := mat.NewVecDense(3, []float64{
		math.Cos(phi2) * math.Cos(theta2),
		math.Sin(phi2) * math.Cos(theta2),
		math.Sin(theta2),
	})
	a1 := mat.NewVecDense(3, []float64{
		-math.Sin(phi1),
		math.Cos(phi1),
		0,
	})
	a2 := mat.NewVecDense(3, []float64{
		-math.Sin(phi2),
		math.Cos(phi2),
		0,
	})
	b1 := mat.NewVecDense(3, []float64{
		-math.Cos(phi1) * math.Sin(theta1),
		-math.Sin(phi1) * math.Sin(theta1),
		math.Cos(theta1),
	})
	b2 := mat.NewVecDense(3, []float64{
		-math.Cos(phi2) * math.Sin(theta2),
		-math.Sin(phi2) * math.Sin(theta2),
		math.Cos(theta2),
	})

	ret := make([][]float64, model.ndata)

	for i := 0; i < model.ndata; i++ {
		snap := model.data[i]
		d1 := mat.NewVecDense(3, nil)
		d1.AddScaledVec(d1, snap.p1, a1)
		d1.AddScaledVec(d1, snap.q1, b1)
		d1.AddScaledVec(n1, k, d1)
		t1 := (z0 - model.c1.At(2, 0)) / d1.At(2, 0)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, snap.p2, a2)
		d2.AddScaledVec(d2, snap.q2, b2)
		d2.AddScaledVec(n2, k, d2)
		t2 := (z0 - model.c2.At(2, 0)) / d2.At(2, 0)

		// x1, y1, x2, y2
		ret[i] = []float64{
			model.c1.At(0, 0) + t1*d1.At(0, 0),
			model.c1.At(1, 0) + t1*d1.At(1, 0),
			model.c2.At(0, 0) + t2*d2.At(0, 0),
			model.c2.At(1, 0) + t2*d2.At(1, 0),
		}
	}
	return ret
}

func (m Model) Plot2D(z0 float64, outDir, fileName string) {
	data := m.project(z0, m.params)
	p := plot.New()
	for _, v := range data {
		ploti1, err := plotter.NewLine(plotter.XYs{
			{X: 0, Y: 0},
			{X: v[0], Y: v[1]},
		})
		if err != nil {
			panic(err)
		}
		ploti1.Color = color.RGBA{255, 0, 0, 128}
		p.Add(ploti1)

		ploti2, err := plotter.NewLine(plotter.XYs{
			{X: 0, Y: 1},
			{X: v[2], Y: v[3]},
		})
		if err != nil {
			panic(err)
		}
		ploti2.Color = color.RGBA{0, 255, 0, 128}
		p.Add(ploti2)

		ploti3, err := plotter.NewLine(plotter.XYs{
			{X: v[0], Y: v[1]},
			{X: v[2], Y: v[3]},
		})
		if err != nil {
			panic(err)
		}
		ploti3.Color = color.RGBA{0, 0, 255, 128}
		p.Add(ploti3)
	}

	p.Add(plotter.NewGrid())
	if err := p.Save(vg.Inch*6, vg.Inch*6, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (model Model) GetPointsDistance(z0 float64, params *mat.VecDense) float64 {
	points := model.project(z0, params)
	sum := .0
	for _, p := range points {
		x1, y1, x2, y2 := p[0], p[1], p[2], p[3]
		sum += (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)
	}
	k := params.At(4, 0)
	return sum / (k * k)
}

func MaxInt(nums ...int64) int64 {
	var ret int64 = math.MinInt64
	for _, v := range nums {
		if ret < v {
			ret = v
		}
	}
	return ret
}

func MinInt(nums ...int64) int64 {
	var ret int64 = math.MaxInt64
	for _, v := range nums {
		if ret > v {
			ret = v
		}
	}
	return ret
}
