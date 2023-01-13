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

type Snapshot struct {
	p1, p2, q1, q2 float64
}

type Config struct {
	Phi1, Phi2, K1, K2, L float64
}

type Model struct {
	nparams, ndata int
	params         *mat.VecDense
	config         Config
	Data           []Snapshot
}

func NewModel(tj1, tj2 Trajectory, params *mat.VecDense, config Config) *Model {
	start := maxInt(tj1.start, tj2.start)
	end := minInt(tj1.end, tj2.end)
	if start > end {
		log.Fatal("These Trajectories have no common segments.")
	}
	tr1 := tj1.plots[start : end+1]
	tr2 := tj2.plots[start : end+1]
	m := new(Model)
	m.ndata = int(end - start + 1) // the number of datasets
	m.Data = make([]Snapshot, m.ndata)
	for i := 0; i < m.ndata; i++ {
		p1, q1 := tr1[i][0]/tj1.width, tr1[i][1]/tj1.height
		p2, q2 := tr2[i][0]/tj2.width, tr2[i][1]/tj2.height
		m.Data[i] = Snapshot{
			p1 - 0.5,
			p2 - 0.5,
			(q1 - 0.5) * (tj1.height / tj1.width),
			(q2 - 0.5) * (tj2.height / tj2.width),
		}
	}
	m.params = params
	m.nparams = params.Len()
	m.config = config
	return m
}

func (m *Model) BatchGradientDecent(dp, mu float64, ntrials int) {
	for i := 0; i < ntrials; i++ {
		// Update phi
		phi1, phi2 := m.getPhis()
		m.config.Phi1 += phi1
		m.config.Phi2 += phi2

		inc := mat.NewVecDense(m.nparams, nil)
		// Update theta, z
		for j := 0; j < m.params.Len()-2; j++ {
			inc.SetVec(j, -m.getDiff(j, dp))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		m.params.AddScaledVec(m.params, mu*math.Exp(-4*float64(i)/float64(ntrials)), inc)
		// m.params.AddScaledVec(m.params, mu, inc)
	}
}

func (m Model) Plot(outDir, fileName string) {
	data := m.project(m.params, m.Data)
	p := plot.New()

	frame := m.project(m.params, []Snapshot{
		{-0.5, -0.5, -0.5 / AspectRatio, -0.5 / AspectRatio}, // bottom left
		{+0.5, +0.5, -0.5 / AspectRatio, -0.5 / AspectRatio}, // bottom right
		{+0.5, +0.5, +0., +0.},                               // mid right
		{-0.5, -0.5, +0., +0.},                               // mid left
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
			{X: 0, Y: 0},
			{X: v[0], Y: v[1]},
		})
		if err != nil {
			panic(err)
		}
		ploti1.Color = color.RGBA{0, 255, 255, 0}
		p.Add(ploti1)

		ploti2, err := plotter.NewLine(plotter.XYs{
			{X: 0, Y: -m.config.L},
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
	// p.X.Max = 3
	// p.X.Min = -0.2
	// p.Y.Max = 1
	// p.Y.Min = -2

	if err := p.Save(vg.Inch*30, vg.Inch*30, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (m Model) Convert(snaps []Snapshot) []*mat.VecDense {
	theta1, theta2 := m.params.At(0, 0), m.params.At(1, 0)
	z1, z2 := m.params.At(2, 0), m.params.At(3, 0)

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

	ret := make([]*mat.VecDense, len(snaps))
	c1 := mat.NewVecDense(3, []float64{0, 0, z1})
	c2 := mat.NewVecDense(3, []float64{0, -m.config.L, z2})
	for i, snap := range snaps {
		d1 := mat.NewVecDense(3, nil)
		d1.AddScaledVec(d1, snap.p1, a1)
		d1.AddScaledVec(d1, snap.q1, b1)
		d1.AddScaledVec(n1, m.config.K1, d1)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, snap.p2, a2)
		d2.AddScaledVec(d2, snap.q2, b2)
		d2.AddScaledVec(n2, m.config.K2, d2)

		mata := mat.NewDense(2, 2, []float64{
			mat.Dot(d1, d1), -mat.Dot(d1, d2),
			mat.Dot(d1, d2), -mat.Dot(d2, d2),
		})
		c21 := mat.NewVecDense(3, nil)
		c21.SubVec(c2, c1)
		matb := mat.NewVecDense(2, []float64{
			mat.Dot(c21, d1),
			mat.Dot(c21, d2),
		})
		t := mat.NewVecDense(2, nil)
		matainv := mat.NewDense(2, 2, nil)
		error := matainv.Inverse(mata)
		if error != nil {
			log.Fatal("Inversed matrix does not exist")
		}
		t.MulVec(matainv, matb)
		l1 := mat.NewVecDense(3, nil)
		l1.AddScaledVec(c1, t.At(0, 0), d1)
		l2 := mat.NewVecDense(3, nil)
		l2.AddScaledVec(c2, t.At(1, 0), d2)
		lm := mat.NewVecDense(3, nil)
		lm.AddVec(l1, l2)
		lm.ScaleVec(0.5, lm)
		ret[i] = lm
	}
	return ret
}

func (m Model) PrintParams(blender bool) {
	if !blender {
		// by radian
		fmt.Printf("theta1: %.2f pi, theta2: %.2f pi, ",
			m.params.At(0, 0)/math.Pi,
			m.params.At(1, 0)/math.Pi,
		)
		fmt.Printf("phi1: %.2f pi, phi2: %.2f pi\t",
			m.config.Phi1/math.Pi,
			m.config.Phi2/math.Pi,
		)
	} else {
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg), ",
			m.params.At(0, 0)/math.Pi*180+90,
			m.config.Phi1/math.Pi*180-90,
		)
		fmt.Printf("(%.2f deg, 0.00 deg, %.2f deg)\t",
			m.params.At(1, 0)/math.Pi*180+90,
			m.config.Phi2/math.Pi*180-90,
		)
	}
	fmt.Printf("z1: %.2f, z2: %.2f ",
		m.params.At(2, 0),
		m.params.At(3, 0),
	)
	fmt.Printf("loss: %.5f\n",
		m.getPointsDistance(m.params),
	)
}

func (m Model) getPhis() (float64, float64) { // phi1, phi2
	data := m.project(m.params, []Snapshot{m.Data[0], m.Data[m.ndata-1]})
	l1 := []float64{data[1][0] - data[0][0], data[1][1] - data[0][1]}
	l2 := []float64{data[1][2] - data[0][2], data[1][3] - data[0][3]}
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

func (m *Model) getDiff(i int, dp float64) float64 {
	cv := m.getPointsDistance(m.params)
	nparams := mat.NewVecDense(m.nparams, nil)
	nparams.SetVec(i, dp)
	nparams.AddVec(m.params, nparams)
	nv := m.getPointsDistance(nparams)
	return (nv - cv) / dp
}

func (m *Model) project(params *mat.VecDense, data []Snapshot) [][]float64 {
	theta1, theta2 := params.At(0, 0), params.At(1, 0)
	z1, z2 := params.At(2, 0), params.At(3, 0)

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

	ret := make([][]float64, len(data))

	for i := 0; i < len(data); i++ {
		snap := data[i]
		d1 := mat.NewVecDense(3, nil)
		d1.AddScaledVec(d1, snap.p1, a1)
		d1.AddScaledVec(d1, snap.q1, b1)
		d1.AddScaledVec(n1, m.config.K1, d1)
		t1 := -z1 / d1.At(2, 0)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, snap.p2, a2)
		d2.AddScaledVec(d2, snap.q2, b2)
		d2.AddScaledVec(n2, m.config.K2, d2)
		t2 := -z2 / d2.At(2, 0)

		d1.AddScaledVec(mat.NewVecDense(3, []float64{
			0, 0, z1,
		}), t1, d1)
		d2.AddScaledVec(mat.NewVecDense(3, []float64{
			0, -m.config.L, z2,
		}), t2, d2)

		// x1, y1, x2, y2
		ret[i] = []float64{
			d1.At(0, 0), d1.At(1, 0),
			d2.At(0, 0), d2.At(1, 0),
		}
	}
	return ret
}

func (m Model) getPointsDistance(params *mat.VecDense) float64 {
	points := m.project(params, m.Data)
	sum := .0
	for _, p := range points {
		x1, y1, x2, y2 := p[0], p[1], p[2], p[3]
		sum += (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)
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
