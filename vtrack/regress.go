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
	"gonum.org/v1/plot/vg/draw"
)

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
		p1, q1 := tr1[i][0]/tj1.width, tr1[i][1]/tj1.height
		p2, q2 := tr2[i][0]/tj2.width, tr2[i][1]/tj2.height
		model.data[i] = Snapshot{
			p1 - 0.5,
			p2 - 0.5,
			(q1 - 0.5) * (tj1.height / tj1.width),
			(q2 - 0.5) * (tj2.height / tj2.width),
		}
	}
	model.c1 = mat.NewVecDense(3, []float64{0, 0, 0})
	model.c2 = mat.NewVecDense(3, []float64{0, -1, 0})
	if len(data) > 0 {
		model.params = data[0]
	} else {
		// Substitute with default parameters
		model.params = mat.NewVecDense(6, []float64{
			-0.06 * math.Pi, -0.18 * math.Pi, // theta
			-0.25 * math.Pi, 0.23 * math.Pi, // phi
			1,    // k
			-0.1, // z0
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

func (m Model) PrintParams(blender bool) {
	if !blender {
		// by radian
		fmt.Printf("theta1: %.2f pi, theta2: %.2f pi, ",
			m.params.At(0, 0)/math.Pi,
			m.params.At(1, 0)/math.Pi,
		)
		fmt.Printf("phi1: %.2f pi, phi2: %.2f pi\t",
			m.params.At(2, 0)/math.Pi,
			m.params.At(3, 0)/math.Pi,
		)
	} else {
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg), ",
			m.params.At(0, 0)/math.Pi*180+90,
			m.params.At(2, 0)/math.Pi*180-90,
		)
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg)\t",
			m.params.At(1, 0)/math.Pi*180+90,
			m.params.At(3, 0)/math.Pi*180-90,
		)
	}
	fmt.Printf("k: %.2f, z0: %.4f, ",
		m.params.At(4, 0),
		m.params.At(5, 0),
	)
	fmt.Printf("loss: %.5f\n",
		m.GetPointsDistance(m.params),
	)
}

func (model *Model) NaiveGradientDecent(dp, mu float64, ntrials int) {
	for i := 0; i < ntrials; i++ {
		for j := 0; j < model.nparams-1; j++ {
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
		for j := 0; j < 2; j++ {
			inc.SetVec(j, -model.GetDiff(j, dp))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		model.params.AddScaledVec(model.params, mu*math.Exp(-4*float64(i)/float64(ntrials)), inc)
	}
}

func (model *Model) GetDiff(i int, dp float64) float64 {
	cv := model.GetPointsDistance(model.params)
	nparams := mat.NewVecDense(model.nparams, nil)
	nparams.SetVec(i, dp)
	nparams.AddVec(model.params, nparams)
	nv := model.GetPointsDistance(nparams)
	return (nv - cv) / dp
}

func (model *Model) project(params *mat.VecDense, data []Snapshot) [][]float64 {
	theta1, theta2 := params.At(0, 0), params.At(1, 0)
	phi1, phi2 := params.At(2, 0), params.At(3, 0)
	k := params.At(4, 0)
	z0 := params.At(5, 0)

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
		math.Sin(phi1),
		-math.Cos(phi1),
		0,
	})
	a2 := mat.NewVecDense(3, []float64{
		math.Sin(phi2),
		-math.Cos(phi2),
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

	ret := make([][]float64, len(data))

	for i := 0; i < len(data); i++ {
		snap := data[i]
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

		d1.AddScaledVec(model.c1, t1, d1)
		d2.AddScaledVec(model.c2, t2, d2)

		// x1, y1, x2, y2
		ret[i] = []float64{
			d1.At(0, 0), d1.At(1, 0),
			d2.At(0, 0), d2.At(1, 0),
		}
	}
	return ret
}

func (m Model) Plot2D(outDir, fileName string) {
	data := m.project(m.params, m.data)
	p := plot.New()

	ratio := height / width
	frame := m.project(m.params, []Snapshot{
		{-0.5, -0.5, -0.5 * ratio, -0.5 * ratio}, // bottom left
		{+0.5, +0.5, -0.5 * ratio, -0.5 * ratio}, // bottom right
		{+0.5, +0.5, +0., +0.},                   // mid right
		{-0.5, -0.5, +0., +0.},                   // mid left
	})
	for i := 0; i < len(frame); i++ {
		f1 := frame[i]
		f2 := frame[(i+1)%len(frame)]
		ploti1, err := plotter.NewLine(plotter.XYs{
			{X: f1[0], Y: f1[1]},
			{X: f2[0], Y: f2[1]},
		})
		if err != nil {
			panic(err)
		}
		ploti2, err := plotter.NewLine(plotter.XYs{
			{X: f1[2], Y: f1[3]},
			{X: f2[2], Y: f2[3]},
		})
		if err != nil {
			panic(err)
		}
		if i == 2 {
			ploti1.LineStyle = draw.LineStyle{
				Color: color.RGBA{255, 255, 255, 0},
				Width: 3.,
			}
			ploti2.LineStyle = draw.LineStyle{
				Color: color.RGBA{255, 255, 255, 0},
				Width: 3.,
			}
		} else {
			ploti1.Color = color.RGBA{0, 255, 255, 128}
			ploti2.Color = color.RGBA{255, 0, 255, 128}
		}
		p.Add(ploti1)
		p.Add(ploti2)
	}
	for _, v := range data {
		ploti1, err := plotter.NewLine(plotter.XYs{
			{X: m.c1.At(0, 0), Y: m.c1.At(1, 0)},
			{X: v[0], Y: v[1]},
		})
		if err != nil {
			panic(err)
		}
		ploti1.Color = color.RGBA{0, 255, 255, 0}
		p.Add(ploti1)

		ploti2, err := plotter.NewLine(plotter.XYs{
			{X: m.c2.At(0, 0), Y: m.c2.At(1, 0)},
			{X: v[2], Y: v[3]},
		})
		if err != nil {
			panic(err)
		}
		ploti2.Color = color.RGBA{255, 0, 255, 0}
		p.Add(ploti2)

		ploti3, err := plotter.NewLine(plotter.XYs{
			{X: v[0], Y: v[1]},
			{X: v[2], Y: v[3]},
		})
		if err != nil {
			panic(err)
		}
		ploti3.Color = color.RGBA{255, 255, 255, 0}
		p.Add(ploti3)
	}

	p.Add(plotter.NewGrid())
	p.X.Max = 3
	p.X.Min = -0.2
	p.Y.Max = 1
	p.Y.Min = -2

	if err := p.Save(vg.Inch*30, vg.Inch*30, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (m Model) GetPointsDistance(params *mat.VecDense) float64 {
	points := m.project(params, m.data)
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
