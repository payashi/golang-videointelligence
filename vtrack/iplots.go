package vtrack

import (
	"encoding/json"

	"github.com/payashi/vannotate"
	"gonum.org/v1/gonum/mat"
)

// Integrated Plots
type IPlots struct {
	Loss       float64
	Size       int
	Plots      *mat.Dense
	Start, End int
	i, j       int
	sr1, sr2   vannotate.Series
}

func (ip *IPlots) UnmarshalJSON(b []byte) error {
	ip2 := &struct {
		Loss  float64     `json:"loss"`
		Size  int         `json:"size"`
		Plots [][]float64 `json:"plots"`
		Start int         `json:"start"`
		End   int         `json:"end"`
		I     int         `json:"i"`
		J     int         `json:"j"`
	}{}
	err := json.Unmarshal(b, ip2)
	ip.Loss = ip2.Loss
	ip.Size = ip2.Size
	ip.Start = ip2.Start
	ip.End = ip2.End
	ip.i = ip2.I
	ip.j = ip2.J
	return err
}

func (ip IPlots) MarshalJSON() ([]byte, error) {
	plots := make([][]float64, ip.Size)
	for i := 0; i < ip.Size; i++ {
		plots[i] = make([]float64, 3)
		for j := 0; j < 3; j++ {
			plots[i][j] = ip.Plots.At(i, j)
		}
	}

	v := &struct {
		Loss  float64     `json:"loss"`
		Size  int         `json:"size"`
		Plots [][]float64 `json:"plots"`
		Start int         `json:"start"`
		End   int         `json:"end"`
		I     int         `json:"i"`
		J     int         `json:"j"`
	}{
		Loss:  ip.Loss,
		Size:  ip.Size,
		Plots: plots,
		Start: ip.Start,
		End:   ip.End,
		I:     ip.i,
		J:     ip.j,
	}
	s, err := json.Marshal(v)
	return s, err
}
