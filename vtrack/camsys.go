package vtrack

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"math"
	"os"

	"github.com/payashi/vannotate"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type Config struct {
	K1, K2 float64
	R1, R2 float64 // aspect ratio
	C1, C2 mat.VecDense
}
type TuneConfig struct {
	Dp, Mu, Z0 float64
	Ntrials    int
	Plots      *splots
}

type CameraSystem struct {
	params  *mat.VecDense // theta1, theta2, phi, phi1, phi2
	config  Config
	tconfig TuneConfig
}

func NewCameraSystem(config Config) *CameraSystem {
	cs := new(CameraSystem)
	cs.config = config
	cs.params = mat.NewVecDense(5, []float64{
		-0.5 * math.Pi, -0.5 * math.Pi, // theta1, theta2
		-0.5 * math.Pi, .0, .0, // phi, phi1, phi2
	})
	return cs
}

func (cs CameraSystem) getConfig(cami int) (float64, float64, mat.VecDense) {
	cf := cs.config
	if cami == 0 {
		return cf.R1, cf.K1, cf.C1
	} else {
		return cf.R2, cf.K2, cf.C2
	}
}

func (cs *CameraSystem) Tune(tconfig TuneConfig) {
	cs.tconfig = tconfig
	for i := 0; i < tconfig.Ntrials; i++ {
		inc := mat.NewVecDense(cs.params.Len(), nil)
		// Update theta1, theta2, phi
		for j := 0; j < 3; j++ {
			inc.SetVec(j, -cs.getDiff(j))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		cs.params.AddScaledVec(
			cs.params,
			tconfig.Mu*math.Exp(-4*float64(i)/float64(tconfig.Ntrials)),
			inc,
		)
	}
}

func (cs CameraSystem) plotFrame(p *plot.Plot) {
	frame := []vannotate.ScreenPlot{
		{P: -0.5, Q: -0.5}, // bottom left
		{P: +0.5, Q: -0.5}, // bottom right
		{P: +0.5, Q: 0.},   // mid right
		{P: -0.5, Q: 0.},   // mid left
	}

	// Plot lower part of frame of each camera
	fm1 := cs.project(cs.params, 0, frame)
	fm2 := cs.project(cs.params, 1, frame)
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
	scatter, err := plotter.NewScatter(plotter.XYs{
		{X: cs.config.C1.At(0, 0), Y: cs.config.C1.At(1, 0)},
		{X: cs.config.C2.At(0, 0), Y: cs.config.C2.At(1, 0)},
	})
	if err != nil {
		panic(err)
	}
	p.Add(scatter)
}

func (cs CameraSystem) PlotJoined(filePath string, iplots []IPlots) {
	p := plot.New()
	cs.plotFrame(p)
	for i, iplot := range iplots {
		fmt.Printf("iplots[%d]: %d-%d\n", i, iplot.i, iplot.j)
		for j := 0; j < iplot.Size-1; j++ {
			ploti, err := plotter.NewLine(plotter.XYs{
				{X: iplot.Plots.At(j, 0), Y: iplot.Plots.At(j, 1)},
				{X: iplot.Plots.At(j+1, 0), Y: iplot.Plots.At(j+1, 1)},
			})
			if err != nil {
				panic(err)
			}
			ploti.Color = plotutil.Color(i)
			p.Add(ploti)
		}
	}
	p.Add(plotter.NewGrid())
	p.X.Max = 15
	p.X.Min = 0
	p.Y.Max = 10
	p.Y.Min = -30

	if err := p.Save(vg.Inch*30, vg.Inch*30, filePath); err != nil {
		panic(err)
	}
}

func (cs CameraSystem) Plot(filePath string, srLists ...[]vannotate.Series) {
	p := plot.New()
	cs.plotFrame(p)

	for cami := 0; cami <= 1; cami++ {
		for _, sr := range srLists[cami] {
			plots := sr.Plots[sr.Start : sr.End+1]
			nplots := len(plots)
			m := cs.project(cs.params, cami, plots)

			for j := 0; j < nplots-1; j++ {
				ploti, err := plotter.NewLine(plotter.XYs{
					{X: m.At(j, 0), Y: m.At(j, 1)},
					{X: m.At(j+1, 0), Y: m.At(j+1, 1)},
				})
				if err != nil {
					panic(err)
				}
				ploti.Color = plotutil.Color(cami)
				p.Add(ploti)
			}
		}
	}

	p.Add(plotter.NewGrid())
	p.X.Max = -5
	p.X.Min = +20
	p.Y.Max = +5
	p.Y.Min = -20

	if err := p.Save(vg.Inch*30, vg.Inch*30, filePath); err != nil {
		panic(err)
	}
}

func (cs CameraSystem) PrintUnityParams() {
	theta1, theta2 := cs.params.At(0, 0), cs.params.At(1, 0)
	phi1, phi2 := cs.params.At(3, 0), cs.params.At(4, 0)
	toDegree := 180 / math.Pi
	theta1 *= toDegree
	theta2 *= toDegree
	phi1 *= toDegree
	phi2 *= toDegree

	// Camera1
	fmt.Println("Camera1:")
	fmt.Printf("\tPosition: [%0.4f, %0.4f, %0.4f]\n",
		cs.config.C1.At(0, 0),
		cs.config.C1.At(2, 0),
		cs.config.C1.At(1, 0),
	)
	fmt.Printf("\tRotation: [%0.4f°, %0.4f°, %0.4f°]\n",
		-theta1,
		90-phi1,
		0.,
	)
	fmt.Printf("\tVertical FOV: %0.4f°\n",
		math.Atan(cs.config.K1/2/cs.config.R1)*2*toDegree,
	)
	fmt.Printf("\tAspect Ratio: %0.4f : 1\n",
		cs.config.R1,
	)
	// Camera2
	fmt.Println("Camera2:")
	fmt.Printf("\tPosition: [%0.4f, %0.4f, %0.4f]\n",
		cs.config.C2.At(0, 0),
		cs.config.C2.At(2, 0),
		cs.config.C2.At(1, 0),
	)
	fmt.Printf("\tRotation: [%0.4f°, %0.4f°, %0.4f°]\n",
		-theta2,
		90-phi2,
		0.,
	)
	fmt.Printf("\tVertical FOV: %0.4f°\n",
		math.Atan(cs.config.K2/2/cs.config.R2)*2*toDegree,
	)
	fmt.Printf("\tAspect Ratio: %0.4f : 1\n",
		cs.config.R2,
	)
}

func (m *CameraSystem) getDiff(i int) float64 {
	cv := m.getPointsDistance(m.params)
	nparams := mat.NewVecDense(m.params.Len(), nil)
	nparams.SetVec(i, m.tconfig.Dp)
	nparams.AddVec(m.params, nparams)
	nv := m.getPointsDistance(nparams)
	return (nv - cv) / m.tconfig.Dp
}

func (cs CameraSystem) getPointsDistance(params *mat.VecDense) float64 {
	// adjust phi1, phi2
	phi1, phi2 := cs.getPhis(params)
	params.SetVec(3, phi1)
	params.SetVec(4, phi2)

	m1 := cs.project(params, 0, cs.tconfig.Plots.pl1)
	m2 := cs.project(params, 1, cs.tconfig.Plots.pl2)

	sum := .0
	for i := 0; i < cs.tconfig.Plots.size; i++ {
		d := mat.NewVecDense(3, nil)
		d.SubVec(m1.RowView(i), m2.RowView(i))
		sum += d.Norm(2)
	}
	return sum
}

func (cs *CameraSystem) getPhis(params *mat.VecDense) (float64, float64) {
	// Get 2D plots
	plots := cs.tconfig.Plots
	pl1 := []vannotate.ScreenPlot{plots.pl1[0], plots.pl1[plots.size-1]}
	pl2 := []vannotate.ScreenPlot{plots.pl2[0], plots.pl2[plots.size-1]}
	m1 := cs.project(cs.params, 0, pl1)
	m2 := cs.project(cs.params, 1, pl2)

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

func (cs *CameraSystem) project(params *mat.VecDense, cami int, plots []vannotate.ScreenPlot, args ...float64) *mat.Dense {
	if cami < 0 || 1 < cami {
		panic("cami should be 0 or 1")
	}
	theta, phi := params.At(0+cami, 0), params.At(3+cami, 0)

	var z0 float64
	if len(args) == 0 {
		z0 = cs.tconfig.Z0
	} else {
		z0 = args[0]
	}

	n := mat.NewVecDense(3, []float64{
		math.Cos(phi) * math.Cos(theta),
		math.Sin(phi) * math.Cos(theta),
		math.Sin(theta),
	})
	a := mat.NewVecDense(3, []float64{
		math.Sin(phi),
		-math.Cos(phi),
		0,
	})
	b := mat.NewVecDense(3, []float64{
		-math.Cos(phi) * math.Sin(theta),
		-math.Sin(phi) * math.Sin(theta),
		math.Cos(theta),
	})

	ret := mat.NewDense(len(plots), 3, nil)
	r, k, c := cs.getConfig(cami)
	for i, plot := range plots {
		d := mat.NewVecDense(3, nil)
		d.AddScaledVec(d, plot.P, a)
		d.AddScaledVec(d, plot.Q/r, b)
		d.AddScaledVec(n, k, d)
		t := (z0 - c.At(2, 0)) / d.At(2, 0)

		d.AddScaledVec(&c, t, d)
		ret.SetRow(i, d.RawVector().Data)
	}
	return ret
}

func (cs CameraSystem) MarshalJSON() ([]byte, error) {
	v := &struct {
		Theta1  float64    `json:"theta1"`
		Theta2  float64    `json:"theta2"`
		Phi     float64    `json:"phi"`
		Phi1    float64    `json:"phi1"`
		Phi2    float64    `json:"phi2"`
		K1      float64    `json:"k1"`
		K2      float64    `json:"k2"`
		R1      float64    `json:"r1"`
		R2      float64    `json:"r2"`
		C1      []float64  `json:"c1"`
		C2      []float64  `json:"c2"`
		TConfig TuneConfig `json:"tconfig"`
	}{
		Theta1:  cs.params.At(0, 0),
		Theta2:  cs.params.At(1, 0),
		Phi:     cs.params.At(2, 0),
		Phi1:    cs.params.At(3, 0),
		Phi2:    cs.params.At(4, 0),
		K1:      cs.config.K1,
		K2:      cs.config.K2,
		R1:      cs.config.R1,
		R2:      cs.config.R2,
		C1:      cs.config.C1.RawVector().Data,
		C2:      cs.config.C2.RawVector().Data,
		TConfig: cs.tconfig,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (cs *CameraSystem) UnmarshalJSON(b []byte) error {
	cs2 := &struct {
		Theta1  float64    `json:"theta1"`
		Theta2  float64    `json:"theta2"`
		Phi     float64    `json:"phi"`
		Phi1    float64    `json:"phi1"`
		Phi2    float64    `json:"phi2"`
		K1      float64    `json:"k1"`
		K2      float64    `json:"k2"`
		R1      float64    `json:"r1"`
		R2      float64    `json:"r2"`
		C1      []float64  `json:"c1"`
		C2      []float64  `json:"c2"`
		TConfig TuneConfig `json:"tconfig"`
	}{}
	err := json.Unmarshal(b, cs2)
	cs.params = mat.NewVecDense(5, []float64{
		cs2.Theta1, cs2.Theta2,
		cs2.Phi, cs2.Phi1, cs2.Phi2,
	})
	cs.config = Config{
		K1: cs2.K1,
		K2: cs2.K2,
		R1: cs2.R1,
		R2: cs2.R2,
		C1: *mat.NewVecDense(3, cs2.C1),
		C2: *mat.NewVecDense(3, cs2.C2),
	}
	cs.tconfig = cs2.TConfig
	return err
}

func LoadCameraSystem(filePath string) (*CameraSystem, error) {
	file, err := os.Open(filePath)
	cs := &CameraSystem{}
	if err != nil {
		return cs, err
	}

	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, cs); err != nil {
		panic(err)
	}
	return cs, nil
}
