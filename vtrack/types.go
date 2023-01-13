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
	Phi1, Phi2, K1, K2 float64
	C1, C2             mat.VecDense
}

type Model struct {
	params *mat.VecDense
	config Config
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
	loss  float64
	size  int
	plots *mat.Dense
	i, j  int
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
		Theta1 float64   `json:"theta1"`
		Theta2 float64   `json:"theta2"`
		Phi1   float64   `json:"phi1"`
		Phi2   float64   `json:"phi2"`
		K1     float64   `json:"k1"`
		K2     float64   `json:"k2"`
		C1     []float64 `json:"c1"`
		C2     []float64 `json:"c2"`
	}{
		Theta1: m.params.At(0, 0),
		Theta2: m.params.At(1, 0),
		Phi1: m.config.Phi1,
		Phi2: m.config.Phi2,
		K1: m.config.K1,
		K2: m.config.K2,
		C1: m.config.C1.RawVector().Data,
		C2: m.config.C2.RawVector().Data,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (m *Model) UnmarshalJSON(b []byte) error {
	m2 := &struct {
		Theta1 float64   `json:"theta1"`
		Theta2 float64   `json:"theta2"`
		Phi1   float64   `json:"phi1"`
		Phi2   float64   `json:"phi2"`
		K1     float64   `json:"k1"`
		K2     float64   `json:"k2"`
		C1     []float64 `json:"c1"`
		C2     []float64 `json:"c2"`
	}{}
	err := json.Unmarshal(b, m2)
	m.params = mat.NewVecDense(2, []float64{m2.Theta1, m2.Theta2})
	m.config = Config{
		Phi1: m2.Phi1,
		Phi2: m2.Phi2,
		K1: m2.K1,
		K2: m2.K2,
		C1: *mat.NewVecDense(3, m2.C1),
		C2: *mat.NewVecDense(3, m2.C2),
	}
	return err
}
