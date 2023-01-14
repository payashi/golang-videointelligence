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
	R1, R2 float64 // aspect ratio
	K1, K2 float64
	C1, C2 mat.VecDense
}
type TuneConfig struct {
	Dp, Mu, Z0 float64
	Ntrials    int
	plots      *splots
}

type CameraSystem struct {
	params  *mat.VecDense // theta1, theta2, phi, phi1, phi2
	config  Config
	tconfig TuneConfig
}

func NewCameraSystem(config Config) *CameraSystem {
	m := new(CameraSystem)
	m.config = config
	m.params = mat.NewVecDense(5, []float64{
		-0.01 * math.Pi, -0.01 * math.Pi, // theta1, theta2
		-0.5 * math.Pi, .0, .0, // phi, phi1, phi2
	})
	return m
}

func (m *CameraSystem) Tune(tconfig TuneConfig) {
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

func (cs CameraSystem) Plot(outDir, fileName string, srList ...[]vannotate.Series) {
	p := plot.New()

	// Plot lower part of frame of each camera
	fm1, fm2 := cs.project(cs.params, &splots{
		size: 4,
		pl1: []vannotate.ScreenPlot{
			{-0.5, -0.5 / cs.config.R1}, // bottom left
			{+0.5, -0.5 / cs.config.R1}, // bottom right
			{+0.5, 0.},                  // mid right
			{-0.5, 0.},                  // mid left
		},
		pl2: []vannotate.ScreenPlot{
			{-0.5, -0.5 / cs.config.R2}, // bottom left
			{+0.5, -0.5 / cs.config.R2}, // bottom right
			{+0.5, 0.},                  // mid right
			{-0.5, 0.},                  // mid left
		},
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

	for i, sr := range srList {
		sr1, sr2 := sr[0], sr[1]
		plots, err := NewSyncedPlots(sr1, sr2)
		if err != nil {
			panic(err)
		}
		m1, m2 := cs.project(cs.params, plots)
		//
		linei1, _, err := plotter.NewLinePoints(plotter.XYs{
			{X: m1.At(0, 0), Y: m1.At(0, 1)},
			{X: cs.config.C1.At(0, 0), Y: cs.config.C1.At(1, 0)},
			{X: m1.At(plots.size-1, 0), Y: m1.At(plots.size-1, 1)},
		})
		if err != nil {
			panic(err)
		}
		linei1.Color = color.Black
		p.Add(linei1)
		//
		linei2, _, err := plotter.NewLinePoints(plotter.XYs{
			{X: m2.At(0, 0), Y: m2.At(0, 1)},
			{X: cs.config.C1.At(0, 0), Y: cs.config.C1.At(1, 0)},
			{X: m2.At(plots.size-1, 0), Y: m2.At(plots.size-1, 1)},
		})
		if err != nil {
			panic(err)
		}
		linei2.Color = color.Black
		p.Add(linei2)
		for j := 0; j < plots.size-1; j++ {
			ploti1, err := plotter.NewLine(plotter.XYs{
				{X: m1.At(i, 0), Y: m1.At(i, 1)},
				{X: m1.At(i+1, 0), Y: m1.At(i+1, 1)},
			})
			if err != nil {
				panic(err)
			}
			ploti1.Color = plotutil.Color(i)
			//
			ploti2, err := plotter.NewLine(plotter.XYs{
				{X: m2.At(i, 0), Y: m2.At(i, 1)},
				{X: m2.At(i+1, 0), Y: m2.At(i+1, 1)},
			})
			if err != nil {
				panic(err)
			}
			ploti2.Color = plotutil.Color(i)
		}
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

func (m CameraSystem) PrintUnityParams() {
	theta1, theta2 := m.params.At(0, 0), m.params.At(1, 0)
	phi1, phi2 := m.params.At(3, 0), m.params.At(4, 0)
	toDegree := 180 / math.Pi
	theta1 *= toDegree
	theta2 *= toDegree
	phi1 *= toDegree
	phi2 *= toDegree

	// Camera1
	fmt.Println("Camera1:")
	fmt.Printf("\tPosition: [%0.4f, %0.4f, %0.4f]\n",
		m.config.C1.At(0, 0),
		m.config.C1.At(2, 0),
		m.config.C1.At(1, 0),
	)
	fmt.Printf("\tRotation: [%0.4f°, %0.4f°, %0.4f°]\n",
		-theta1,
		90-phi1,
		0.,
	)
	fmt.Printf("\tVertical FOV: %0.4f°\n",
		math.Atan(m.config.K1/2/m.config.R1)*2*toDegree,
	)
	fmt.Printf("\tAspect Ratio: %0.4f : 1\n",
		m.config.R1,
	)
	// Camera2
	fmt.Println("Camera2:")
	fmt.Printf("\tPosition: [%0.4f, %0.4f, %0.4f]\n",
		m.config.C2.At(0, 0),
		m.config.C2.At(2, 0),
		m.config.C2.At(1, 0),
	)
	fmt.Printf("\tRotation: [%0.4f°, %0.4f°, %0.4f°]\n",
		-theta2,
		90-phi2,
		0.,
	)
	fmt.Printf("\tVertical FOV: %0.4f°\n",
		math.Atan(m.config.K2/2/m.config.R2)*2*toDegree,
	)
	fmt.Printf("\tAspect Ratio: %0.4f : 1\n",
		m.config.R2,
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

func (m CameraSystem) getPointsDistance(params *mat.VecDense) float64 {
	// adjust phi1, phi2
	phi1, phi2 := m.getPhis(params)
	params.SetVec(3, phi1)
	params.SetVec(4, phi2)

	m1, m2 := m.project(params, m.tconfig.plots)

	sum := .0
	for i := 0; i < m.tconfig.plots.size; i++ {
		d := mat.NewVecDense(3, nil)
		d.SubVec(m1.RowView(i), m2.RowView(i))
		sum += d.Norm(2)
	}
	return sum
}

func (m *CameraSystem) getPhis(params *mat.VecDense) (float64, float64) {
	plots := m.tconfig.plots
	m1, m2 := m.project(params, &splots{
		size: 2,
		pl1:  []vannotate.ScreenPlot{plots.pl1[0], plots.pl1[plots.size-1]},
		pl2:  []vannotate.ScreenPlot{plots.pl2[0], plots.pl2[plots.size-1]},
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

func (m *CameraSystem) project(params *mat.VecDense, plots *splots, args ...float64) (*mat.Dense, *mat.Dense) {
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
		d1.AddScaledVec(d1, pl1.P, a1)
		d1.AddScaledVec(d1, pl1.Q/m.config.R1, b1)
		d1.AddScaledVec(n1, m.config.K1, d1)
		t1 := (z0 - m.config.C1.At(2, 0)) / d1.At(2, 0)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, pl2.P, a2)
		d2.AddScaledVec(d2, pl2.Q/m.config.R2, b2)
		d2.AddScaledVec(n2, m.config.K2, d2)
		t2 := (z0 - m.config.C2.At(2, 0)) / d2.At(2, 0)

		d1.AddScaledVec(&m.config.C1, t1, d1)
		d2.AddScaledVec(&m.config.C2, t2, d2)

		ret1.SetRow(i, d1.RawVector().Data)
		ret2.SetRow(i, d2.RawVector().Data)
	}
	return ret1, ret2
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
		C1:      cs.config.C1.RawVector().Data,
		C2:      cs.config.C2.RawVector().Data,
		TConfig: cs.tconfig,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (cs *CameraSystem) UnmarshalJSON(b []byte) error {
	m2 := &struct {
		Theta1  float64    `json:"theta1"`
		Theta2  float64    `json:"theta2"`
		Phi     float64    `json:"phi"`
		Phi1    float64    `json:"phi1"`
		Phi2    float64    `json:"phi2"`
		K1      float64    `json:"k1"`
		K2      float64    `json:"k2"`
		C1      []float64  `json:"c1"`
		C2      []float64  `json:"c2"`
		TConfig TuneConfig `json:"tconfig"`
	}{}
	err := json.Unmarshal(b, m2)
	cs.params = mat.NewVecDense(5, []float64{
		m2.Theta1, m2.Theta2,
		m2.Phi, m2.Phi1, m2.Phi2,
	})
	cs.config = Config{
		K1: m2.K1,
		K2: m2.K2,
		C1: *mat.NewVecDense(3, m2.C1),
		C2: *mat.NewVecDense(3, m2.C2),
	}
	cs.tconfig = m2.TConfig
	return err
}

func GetCameraSystem(outDir, fileName string, sr1, sr2 vannotate.Series, config Config, tconfig TuneConfig) *CameraSystem {
	filePath := fmt.Sprintf("%s/%s.json", outDir, fileName)
	// If file is found, read and unmarshal it
	if file, err := os.Open(filePath); err == nil {
		defer file.Close()
		b, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		var ret CameraSystem
		if err := json.Unmarshal(b, &ret); err != nil {
			panic(err)
		}
		return &ret
	} else {
		fmt.Printf("Tuning %s...\n", fileName)
		plots, _ := NewSyncedPlots(sr1, sr2)
		tconfig.plots = plots

		ret := NewCameraSystem(config)
		ret.Tune(tconfig)
		ret.Plot(outDir, "after")

		// Save on local
		newFile, err := json.MarshalIndent(ret, "", "\t")
		if err != nil {
			panic(err)
		}
		_ = ioutil.WriteFile(filePath, newFile, 0644)
		return ret
	}
}
