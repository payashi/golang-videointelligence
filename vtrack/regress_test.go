package vtrack

import (
	"testing"
)

func TestRegress(t *testing.T) {
	tj1 := Trajectory{
		conf: 1,
		plots: [][]float64{
			{1280, 360},
			{960, 360},
			{640, 360},
			{320, 360},
			{0, 360},
		},
		start:  0,
		end:    99,
		width:  1280,
		height: 720,
	}
	tj2 := Trajectory{
		conf: 1,
		plots: [][]float64{
			{0, 360},
			{320, 360},
			{640, 360},
			{960, 360},
			{1280, 360},
		},
		start:  0,
		end:    99,
		width:  1280,
		height: 720,
	}
	Regress(tj1, tj2)
}
