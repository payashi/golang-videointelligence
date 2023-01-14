package vtrack

import (
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

// func (tdp IPlots) MarshalJSON() ([]byte, error) {
// 	plots := make([][]float64, tdp.size)
// 	for i := 0; i < tdp.size; i++ {
// 		plots[i] = []float64{
// 			tdp.plots.At(i, 0),
// 			tdp.plots.At(i, 1),
// 			tdp.plots.At(i, 2),
// 		}
// 	}
// 	v := struct {
// 		Loss  float64     `json:"loss"`
// 		Size  int         `json:"size"`
// 		I     int         `json:"i"`
// 		J     int         `json:"j"`
// 		Start int         `json:"start"`
// 		End   int         `json:"end"`
// 		Plots [][]float64 `json:"plots"`
// 	}{
// 		Loss:  tdp.loss,
// 		Size:  tdp.size,
// 		I:     tdp.i,
// 		J:     tdp.j,
// 		Start: tdp.start,
// 		End:   tdp.end,
// 		Plots: plots,
// 	}
// 	s, err := json.Marshal(v)
// 	return s, err
// }

// func (tdp *IPlots) UnmarshalJSON(b []byte) error {
// 	tdp2 := &struct {
// 		Loss  float64     `json:"loss"`
// 		Size  int         `json:"size"`
// 		I     int         `json:"i"`
// 		J     int         `json:"j"`
// 		Start int         `json:"start"`
// 		End   int         `json:"end"`
// 		Plots [][]float64 `json:"plots"`
// 	}{}
// 	err := json.Unmarshal(b, tdp2)
// 	tdp.loss = tdp2.Loss
// 	tdp.size = tdp2.Size
// 	tdp.i = tdp2.I
// 	tdp.j = tdp2.J
// 	tdp.start = tdp2.Start
// 	tdp.end = tdp2.End
// 	plots := mat.NewDense(len(tdp2.Plots), 3, nil)
// 	for i, v := range tdp2.Plots {
// 		plots.SetRow(i, v)
// 	}
// 	tdp.plots = plots
// 	return err
// }
