package vtrack

import (
	"encoding/json"

	"gonum.org/v1/gonum/mat"
)

const AspectRatio float64 = 16. / 9.
const MaxDur int = 601 // time range

type Snapshot struct {
	sc1, sc2 ScreenCoordinate
}

type ScreenCoordinate struct {
	p, q float64
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

type Trajectory struct {
	conf          float32
	plots         []ScreenCoordinate
	start, end    int64
	length        float64
	width, height float64
}

type AnnotationResults struct {
	trajectories []Trajectory
}

func (sc ScreenCoordinate) MarshalJSON() ([]byte, error) {
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

func (sc *ScreenCoordinate) UnmarshalJSON(b []byte) error {
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
		Conf   float32            `json:"conf"`
		Plots  []ScreenCoordinate `json:"plots"`
		Start  int64              `json:"start"`
		End    int64              `json:"end"`
		Length float64            `json:"length"`
		Width  float64            `json:"width"`
		Height float64            `json:"height"`
	}{
		Conf:   tj.conf,
		Plots:  tj.plots,
		Start:  tj.start,
		End:    tj.end,
		Length: tj.length,
		Width:  tj.width,
		Height: tj.height,
	}
	s, err := json.Marshal(v)
	return s, err
}

func (tj *Trajectory) UnmarshalJSON(b []byte) error {
	tj2 := &struct {
		Conf   float32            `json:"conf"`
		Plots  []ScreenCoordinate `json:"plots"`
		Start  int64              `json:"start"`
		End    int64              `json:"end"`
		Length float64            `json:"length"`
		Width  float64            `json:"width"`
		Height float64            `json:"height"`
	}{}
	err := json.Unmarshal(b, tj2)
	tj.conf = tj2.Conf
	tj.plots = tj2.Plots
	tj.start = tj2.Start
	tj.end = tj2.End
	tj.length = tj2.Length
	tj.width = tj2.Width
	tj.height = tj2.Height
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
