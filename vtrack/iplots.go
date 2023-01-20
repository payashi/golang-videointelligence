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
		Start int         `json:"start"`
		End   int         `json:"end"`
		I     int         `json:"i"`
		J     int         `json:"j"`
		Plots [][]float64 `json:"plots"`
	}{}
	err := json.Unmarshal(b, ip2)
	ip.Loss = ip2.Loss
	ip.Size = ip2.Size
	ip.Start = ip2.Start
	ip.End = ip2.End
	ip.i = ip2.I
	ip.j = ip2.J
	ip.Plots = mat.NewDense(len(ip2.Plots), 3, nil)
	for i := 0; i < len(ip2.Plots); i++ {
		ip.Plots.SetRow(i, ip2.Plots[i])
	}
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
		Start int         `json:"start"`
		End   int         `json:"end"`
		I     int         `json:"i"`
		J     int         `json:"j"`
		Plots [][]float64 `json:"plots"`
	}{
		Loss:  ip.Loss,
		Size:  ip.Size,
		Start: ip.Start,
		End:   ip.End,
		I:     ip.i,
		J:     ip.j,
		Plots: plots,
	}
	s, err := json.Marshal(v)
	return s, err
}
