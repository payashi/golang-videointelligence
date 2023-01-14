package vtrack

import (
	"errors"
	"fmt"
	"image/color"
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func NewModel(config Config) *Model {
	m := new(Model)
	m.config = config
	m.params = mat.NewVecDense(5, []float64{
		-0.01 * math.Pi, -0.01 * math.Pi, // theta1, theta2
		-0.5 * math.Pi, .0, .0, // phi, phi1, phi2
	})
	return m
}

func NewSyncedPlots(tj1, tj2 Trajectory) (*SyncedPlots, error) {
	start := maxInt(tj1.start, tj2.start)
	end := minInt(tj1.end, tj2.end)
	if start > end {
		return nil, errors.New("syncedplots: no overwrapped span")
	}
	pl1 := tj1.plots[start : end+1]
	pl2 := tj2.plots[start : end+1]
	return &SyncedPlots{
		size:  len(pl1),
		start: int(start),
		end:   int(end),
		pl1:   pl1,
		pl2:   pl2,
	}, nil
}

func (m *Model) Tune(tconfig TuneConfig) {
	m.tconfig = tconfig
	for i := 0; i < tconfig.Ntrials; i++ {
		inc := mat.NewVecDense(m.params.Len(), nil)
		// Update theta1, theta2, phi
		for j := 0; j < 3; j++ {
			inc.SetVec(j, -m.getDiff(j))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		m.params.AddScaledVec(
			m.params,
			tconfig.Mu*math.Exp(-4*float64(i)/float64(tconfig.Ntrials)),
			inc,
		)
	}
}

func (m Model) Plot(outDir, fileName string) {
	p := plot.New()

	corners := []ScreenPlot{
		{-0.5, -0.5 / AspectRatio}, // bottom left
		{+0.5, -0.5 / AspectRatio}, // bottom right
		{+0.5, 0.},                 // mid right
		{-0.5, 0.},                 // mid left
	}

	fm1, fm2 := m.project(m.params, &SyncedPlots{
		size: len(corners),
		pl1:  corners,
		pl2:  corners,
	})
	for i := 0; i < 4; i++ {
		ni := (i + 1) % 4
		ploti1, err := plotter.NewLine(plotter.XYs{
			{X: fm1.At(i, 0), Y: fm1.At(i, 1)},
			{X: fm1.At(ni, 0), Y: fm1.At(ni, 1)},
		})
		if err != nil {
			panic(err)
		}
		ploti2, err := plotter.NewLine(plotter.XYs{
			{X: fm2.At(i, 0), Y: fm2.At(i, 1)},
			{X: fm2.At(ni, 0), Y: fm2.At(ni, 1)},
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

	m1, m2 := m.project(m.params, m.tconfig.Plots)
	for i := 0; i < m.tconfig.Plots.size; i++ {
		ploti1, err := plotter.NewLine(plotter.XYs{
			{X: m.config.C1.At(0, 0), Y: m.config.C1.At(1, 0)},
			{X: m1.At(i, 0), Y: m1.At(i, 1)},
		})
		if err != nil {
			panic(err)
		}
		ploti1.Color = color.RGBA{0, 255, 255, 0}
		p.Add(ploti1)

		ploti2, err := plotter.NewLine(plotter.XYs{
			{X: m.config.C2.At(0, 0), Y: m.config.C2.At(1, 0)},
			{X: m2.At(i, 0), Y: m2.At(i, 1)},
		})
		if err != nil {
			panic(err)
		}
		ploti2.Color = color.RGBA{255, 0, 255, 0}
		p.Add(ploti2)

		ploti3, err := plotter.NewLine(plotter.XYs{
			{X: m1.At(i, 0), Y: m1.At(i, 1)},
			{X: m2.At(i, 0), Y: m2.At(i, 1)},
		})
		if err != nil {
			panic(err)
		}
		ploti3.Color = color.RGBA{255, 255, 255, 0}
		p.Add(ploti3)
	}

	p.Add(plotter.NewGrid())
	// p.X.Max = 10
	// p.X.Min = 0
	// p.Y.Max = 0
	// p.Y.Min = -20

	if err := p.Save(vg.Inch*30, vg.Inch*30, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (m Model) PrintParams(blender bool) {
	theta1, theta2 := m.params.At(0, 0), m.params.At(1, 0)
	phi1, phi2 := m.params.At(3, 0), m.params.At(4, 0)
	if !blender {
		// by radian
		fmt.Printf("theta1: %.2f pi, theta2: %.2f pi, ",
			theta1/math.Pi,
			theta2/math.Pi,
		)
		fmt.Printf("phi1: %.2f pi, phi2: %.2f pi\n",
			phi1/math.Pi,
			phi2/math.Pi,
		)
	} else {
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg), ",
			theta1/math.Pi*180+90,
			phi1/math.Pi*180-90,
		)
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg)\n",
			theta2/math.Pi*180+90,
			phi2/math.Pi*180-90,
		)
	}
}

func (m *Model) getDiff(i int) float64 {
	cv := m.getPointsDistance(m.params)
	nparams := mat.NewVecDense(m.params.Len(), nil)
	nparams.SetVec(i, m.tconfig.Dp)
	nparams.AddVec(m.params, nparams)
	nv := m.getPointsDistance(nparams)
	return (nv - cv) / m.tconfig.Dp
}
func (m Model) getPointsDistance(params *mat.VecDense) float64 {
	// adjust phi1, phi2
	phi1, phi2 := m.getPhis(params)
	params.SetVec(3, phi1)
	params.SetVec(4, phi2)

	m1, m2 := m.project(params, m.tconfig.Plots)

	sum := .0
	for i := 0; i < m.tconfig.Plots.size; i++ {
		d := mat.NewVecDense(3, nil)
		d.SubVec(m1.RowView(i), m2.RowView(i))
		sum += d.Norm(2)
	}
	return sum
}
func (m *Model) getPhis(params *mat.VecDense) (float64, float64) {
	plots := m.tconfig.Plots
	m1, m2 := m.project(params, &SyncedPlots{
		size: 2,
		pl1:  []ScreenPlot{plots.pl1[0], plots.pl1[plots.size-1]},
		pl2:  []ScreenPlot{plots.pl2[0], plots.pl2[plots.size-1]},
	})
	d1 := mat.NewVecDense(3, nil)
	d2 := mat.NewVecDense(3, nil)
	d1.SubVec(m1.RowView(1), m1.RowView(0))
	d2.SubVec(m2.RowView(1), m2.RowView(0))
	t1 := math.Atan2(d1.At(1, 0), d1.At(0, 0))
	t2 := math.Atan2(d2.At(1, 0), d2.At(0, 0))
	phi := params.At(2, 0)
	phi1 := params.At(3, 0) + phi - t1
	phi2 := params.At(4, 0) + phi - t2
	return phi1, phi2
}

func (m *Model) project(params *mat.VecDense, plots *SyncedPlots, args ...float64) (*mat.Dense, *mat.Dense) {
	theta1, theta2 := params.At(0, 0), params.At(1, 0)
	phi1, phi2 := params.At(3, 0), params.At(4, 0)
	var z0 float64
	if len(args) == 0 {
		z0 = m.tconfig.Z0
	} else {
		z0 = args[0]
	}

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

	ret1 := mat.NewDense(plots.size, 3, nil)
	ret2 := mat.NewDense(plots.size, 3, nil)

	for i := 0; i < plots.size; i++ {
		pl1 := plots.pl1[i]
		pl2 := plots.pl2[i]

		d1 := mat.NewVecDense(3, nil)
		d1.AddScaledVec(d1, pl1.p, a1)
		d1.AddScaledVec(d1, pl1.q, b1)
		d1.AddScaledVec(n1, m.config.K1, d1)
		t1 := (z0 - m.config.C1.At(2, 0)) / d1.At(2, 0)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, pl2.p, a2)
		d2.AddScaledVec(d2, pl2.q, b2)
		d2.AddScaledVec(n2, m.config.K2, d2)
		t2 := (z0 - m.config.C2.At(2, 0)) / d2.At(2, 0)

		d1.AddScaledVec(&m.config.C1, t1, d1)
		d2.AddScaledVec(&m.config.C2, t2, d2)

		ret1.SetRow(i, d1.RawVector().Data)
		ret2.SetRow(i, d2.RawVector().Data)
	}
	return ret1, ret2
}

func maxInt(nums ...int64) int64 {
	var ret int64 = math.MinInt64
	for _, v := range nums {
		if ret < v {
			ret = v
		}
	}
	return ret
}

func minInt(nums ...int64) int64 {
	var ret int64 = math.MaxInt64
	for _, v := range nums {
		if ret > v {
			ret = v
		}
	}
	return ret
}
