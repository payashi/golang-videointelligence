package vtrack

import (
	"encoding/json"

	"gonum.org/v1/gonum/mat"
)

const AspectRatio float64 = 16. / 9.
const MaxDur int = 601 // time range

type SyncedPlots struct {
	pl1, pl2   []ScreenPlot
	start, end int
	size       int
}

type ScreenPlot struct {
	p, q float64
}

type Config struct {
	K1, K2 float64
	C1, C2 mat.VecDense
}
type TuneConfig struct {
	Dp, Mu, Z0 float64
	Ntrials    int
	Plots      *SyncedPlots
}

type Model struct {
	params  *mat.VecDense
	config  Config
	tconfig TuneConfig
}

type Trajectory struct {
	conf       float32
	plots      []ScreenPlot
	start, end int64
	length     float64
}

type AnnotationResults struct {
	Trajectories []Trajectory
}

type ThreeDimensionalPlots struct {
	loss       float64
	size       int
	plots      *mat.Dense
	i, j       int
	start, end int
}

func (sc ScreenPlot) MarshalJSON() ([]byte, error) {
	v := &struct {
		P float64 `json:"p"`
		Q float64 `json:"q"`
	}{
		P: sc.p,
		Q: sc.q,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (sc *ScreenPlot) UnmarshalJSON(b []byte) error {
	sc2 := &struct {
		P float64 `json:"p"`
		Q float64 `json:"q"`
	}{}
	err := json.Unmarshal(b, sc2)
	sc.p = sc2.P
	sc.q = sc2.Q
	return err
}

func (tj Trajectory) MarshalJSON() ([]byte, error) {
	v := &struct {
		Conf   float32      `json:"conf"`
		Plots  []ScreenPlot `json:"plots"`
		Start  int64        `json:"start"`
		End    int64        `json:"end"`
		Length float64      `json:"length"`
	}{
		Conf:   tj.conf,
		Plots:  tj.plots,
		Start:  tj.start,
		End:    tj.end,
		Length: tj.length,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (tj *Trajectory) UnmarshalJSON(b []byte) error {
	tj2 := &struct {
		Conf   float32      `json:"conf"`
		Plots  []ScreenPlot `json:"plots"`
		Start  int64        `json:"start"`
		End    int64        `json:"end"`
		Length float64      `json:"length"`
	}{}
	err := json.Unmarshal(b, tj2)
	tj.conf = tj2.Conf
	tj.plots = tj2.Plots
	tj.start = tj2.Start
	tj.end = tj2.End
	tj.length = tj2.Length
	return err
}

func (m Model) MarshalJSON() ([]byte, error) {
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
		Theta1:  m.params.At(0, 0),
		Theta2:  m.params.At(1, 0),
		Phi:     m.params.At(2, 0),
		Phi1:    m.params.At(3, 0),
		Phi2:    m.params.At(4, 0),
		K1:      m.config.K1,
		K2:      m.config.K2,
		C1:      m.config.C1.RawVector().Data,
		C2:      m.config.C2.RawVector().Data,
		TConfig: m.tconfig,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (m *Model) UnmarshalJSON(b []byte) error {
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
	m.params = mat.NewVecDense(5, []float64{
		m2.Theta1, m2.Theta2,
		m2.Phi, m2.Phi1, m2.Phi2,
	})
	m.config = Config{
		K1: m2.K1,
		K2: m2.K2,
		C1: *mat.NewVecDense(3, m2.C1),
		C2: *mat.NewVecDense(3, m2.C2),
	}
	m.tconfig = m2.TConfig
	return err
}

type xyz struct {
	X, Y, Z float64
}

func (tdp ThreeDimensionalPlots) MarshalJSON() ([]byte, error) {
	plots := make([]xyz, tdp.size)
	for i := 0; i < tdp.size; i++ {
		plots[i] = xyz{
			X: tdp.plots.At(i, 0),
			Y: tdp.plots.At(i, 1),
			Z: tdp.plots.At(i, 2),
		}
	}
	v := struct {
		Loss  float64 `json:"loss"`
		Size  int     `json:"size"`
		I     int     `json:"i"`
		J     int     `json:"j"`
		Start int     `json:"start"`
		End   int     `json:"end"`
		Plots []xyz   `json:"plots"`
	}{
		Loss:  tdp.loss,
		Size:  tdp.size,
		I:     tdp.i,
		J:     tdp.j,
		Start: tdp.start,
		End:   tdp.end,
		Plots: plots,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (tdp *ThreeDimensionalPlots) UnmarshalJSON(b []byte) error {
	tdp2 := &struct {
		Loss  float64 `json:"loss"`
		Size  int     `json:"size"`
		I     int     `json:"i"`
		J     int     `json:"j"`
		Start int     `json:"start"`
		End   int     `json:"end"`
		Plots []xyz   `json:"plots"`
	}{}
	err := json.Unmarshal(b, tdp2)
	tdp.loss = tdp2.Loss
	tdp.size = tdp2.Size
	tdp.i = tdp2.I
	tdp.j = tdp2.J
	tdp.start = tdp2.Start
	tdp.end = tdp2.End
	plots := mat.NewDense(len(tdp2.Plots), 3, nil)
	for i, v := range tdp2.Plots {
		plots.SetRow(i, []float64{v.X, v.Y, v.Z})
	}
	tdp.plots = plots
	return err
}
