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
	m.params = mat.NewVecDense(2, []float64{
		-0.01 * math.Pi, -0.01 * math.Pi, // theta1, theta2
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
		size: len(pl1),
		pl1:  pl1,
		pl2:  pl2,
	}, nil
}

func (m *Model) Tune(
	plots *SyncedPlots,
	dp, mu float64,
	ntrials int,
) {
	for i := 0; i < ntrials; i++ {
		// Update phi
		phi1, phi2 := m.getPhis(plots)
		m.config.Phi1 += phi1
		m.config.Phi2 += phi2

		inc := mat.NewVecDense(m.params.Len(), nil)
		// Update theta
		for j := 0; j < m.params.Len(); j++ {
			inc.SetVec(j, -m.getDiff(j, dp, plots))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		m.params.AddScaledVec(m.params, mu*math.Exp(-4*float64(i)/float64(ntrials)), inc)
	}
}

func (m Model) Plot(outDir, fileName string, plots *SyncedPlots) {
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

	m1, m2 := m.project(m.params, plots)
	for i := 0; i < plots.size; i++ {
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
	p.X.Max = 10
	p.X.Min = 0
	p.Y.Max = 0
	p.Y.Min = -20

	if err := p.Save(vg.Inch*30, vg.Inch*30, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (m Model) PrintParams(blender bool) {
	if !blender {
		// by radian
		fmt.Printf("theta1: %.2f pi, theta2: %.2f pi, ",
			m.params.At(0, 0)/math.Pi,
			m.params.At(1, 0)/math.Pi,
		)
		fmt.Printf("phi1: %.2f pi, phi2: %.2f pi\n",
			m.config.Phi1/math.Pi,
			m.config.Phi2/math.Pi,
		)
	} else {
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg), ",
			m.params.At(0, 0)/math.Pi*180+90,
			m.config.Phi1/math.Pi*180-90,
		)
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg)\n",
			m.params.At(1, 0)/math.Pi*180+90,
			m.config.Phi2/math.Pi*180-90,
		)
	}
}

func (m Model) getPhis(plots *SyncedPlots) (float64, float64) { // phi1, phi2
	m1, m2 := m.project(m.params, &SyncedPlots{
		size: 2,
		pl1:  []ScreenPlot{plots.pl1[0], plots.pl1[plots.size-1]},
		pl2:  []ScreenPlot{plots.pl2[0], plots.pl2[plots.size-1]},
	})
	d1 := mat.NewVecDense(3, nil)
	d2 := mat.NewVecDense(3, nil)
	d1.SubVec(m1.RowView(1), m1.RowView(0))
	d2.SubVec(m2.RowView(1), m2.RowView(0))
	l1 := d1.RawVector().Data
	l2 := d2.RawVector().Data
	var phi1, phi2 float64
	if l1[0] > 0 {
		phi1 = -0.5*math.Pi - math.Atan(l1[1]/l1[0])
	} else if l1[0] == 0 {
		if l1[1] > 0 {
			phi1 = math.Pi
		} else {
			phi1 = 0
		}
	} else {
		phi1 = +0.5*math.Pi - math.Atan(l1[1]/l1[0])
	}
	if l2[0] > 0 {
		phi2 = -0.5*math.Pi - math.Atan(l2[1]/l2[0])
	} else if l2[0] == 0 {
		if l2[1] > 0 {
			phi2 = math.Pi
		} else {
			phi2 = 0
		}
	} else {
		phi2 = +0.5*math.Pi - math.Atan(l2[1]/l2[0])
	}
	return phi1, phi2
}

func (m *Model) getDiff(i int, dp float64, plots *SyncedPlots) float64 {
	cv := m.getPointsDistance(m.params, plots)
	nparams := mat.NewVecDense(m.params.Len(), nil)
	nparams.SetVec(i, dp)
	nparams.AddVec(m.params, nparams)
	nv := m.getPointsDistance(nparams, plots)
	return (nv - cv) / dp
}

func (m *Model) project(params *mat.VecDense, plots *SyncedPlots) (*mat.Dense, *mat.Dense) {
	theta1, theta2 := params.At(0, 0), params.At(1, 0)

	n1 := mat.NewVecDense(3, []float64{
		math.Cos(m.config.Phi1) * math.Cos(theta1),
		math.Sin(m.config.Phi1) * math.Cos(theta1),
		math.Sin(theta1),
	})
	n2 := mat.NewVecDense(3, []float64{
		math.Cos(m.config.Phi2) * math.Cos(theta2),
		math.Sin(m.config.Phi2) * math.Cos(theta2),
		math.Sin(theta2),
	})
	a1 := mat.NewVecDense(3, []float64{
		math.Sin(m.config.Phi1),
		-math.Cos(m.config.Phi1),
		0,
	})
	a2 := mat.NewVecDense(3, []float64{
		math.Sin(m.config.Phi2),
		-math.Cos(m.config.Phi2),
		0,
	})
	b1 := mat.NewVecDense(3, []float64{
		-math.Cos(m.config.Phi1) * math.Sin(theta1),
		-math.Sin(m.config.Phi1) * math.Sin(theta1),
		math.Cos(theta1),
	})
	b2 := mat.NewVecDense(3, []float64{
		-math.Cos(m.config.Phi2) * math.Sin(theta2),
		-math.Sin(m.config.Phi2) * math.Sin(theta2),
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
		t1 := -m.config.C1.At(2, 0) / d1.At(2, 0)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, pl2.p, a2)
		d2.AddScaledVec(d2, pl2.q, b2)
		d2.AddScaledVec(n2, m.config.K2, d2)
		t2 := -m.config.C2.At(2, 0) / d2.At(2, 0)

		d1.AddScaledVec(&m.config.C1, t1, d1)
		d2.AddScaledVec(&m.config.C2, t2, d2)

		ret1.SetRow(i, d1.RawVector().Data)
		ret2.SetRow(i, d2.RawVector().Data)
	}
	return ret1, ret2
}

func (m Model) getPointsDistance(params *mat.VecDense, plots *SyncedPlots) float64 {
	m1, m2 := m.project(params, plots)

	sum := .0
	for i := 0; i < plots.size; i++ {
		d := mat.NewVecDense(3, nil)
		d.SubVec(m1.RowView(i), m2.RowView(i))
		sum += d.Norm(2)
	}
	return sum
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
