package vtrack

import (
	"testing"
)

func TestRegress(t *testing.T) {
	tj1 := Trajectory{
		Conf: 1,
		Plots: [][]float64{
			{1280, 360},
			{960, 360},
			{640, 360},
			{320, 360},
			{0, 360},
		},
		Start:  0,
		End:    99,
		Width:  1280,
		Height: 720,
	}
	tj2 := Trajectory{
		Conf: 1,
		Plots: [][]float64{
			{0, 360},
			{320, 360},
			{640, 360},
			{960, 360},
			{1280, 360},
		},
		Start:  0,
		End:    99,
		Width:  1280,
		Height: 720,
	}
	Regress(tj1, tj2)

}
