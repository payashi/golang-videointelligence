package vtrack

import (
	"encoding/json"

	"gonum.org/v1/gonum/mat"
)

const AspectRatio float64 = 16. / 9.
const MaxDur int = 601 // time range

type SyncedPlots struct {
	pl1, pl2 []ScreenPlot
	size     int
}

type ScreenPlot struct {
	p, q float64
}

type Config struct {
	Phi1, Phi2, K1, K2, L float64
}

type Model struct {
	nparams int
	params  *mat.VecDense
	config  Config
	Data    SyncedPlots
}

type Trajectory struct {
	conf       float32
	plots      []ScreenPlot
	start, end int64
	length     float64
}

type AnnotationResults struct {
	trajectories []Trajectory
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

func (ar AnnotationResults) MarshalJSON() ([]byte, error) {
	v := &struct {
		Trajectories []Trajectory `json:"trajectories"`
	}{ar.trajectories}
	s, err := json.Marshal(v)
	return s, err
}

func (ar *AnnotationResults) UnmarshalJSON(b []byte) error {
	ar2 := &struct {
		Trajectories []Trajectory `json:"trajectories"`
	}{}
	err := json.Unmarshal(b, ar2)
	ar.trajectories = ar2.Trajectories
	return err
}
