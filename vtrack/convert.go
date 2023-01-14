package vtrack

import (
	"fmt"
	"sort"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func Plot(outDir, fileName string, tdps []ThreeDimensionalPlots) {
	p := plot.New()
	for i, tdp := range tdps {
		pts := make(plotter.XYs, tdp.size)
		for i := 0; i < tdp.size; i++ {
			pts[i] = plotter.XY{X: tdp.plots.At(i, 0), Y: tdp.plots.At(i, 1)}
		}
		line, _, err := plotter.NewLinePoints(pts)
		if err != nil {
			panic(err)
		}
		line.Color = plotutil.Color(i)
		p.Add(line)
	}
	p.X.Max = 10
	p.X.Min = 0
	p.Y.Max = 0
	p.Y.Min = -20
	p.Add(plotter.NewGrid())

	if err := p.Save(vg.Inch*30, vg.Inch*30, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}

func (m Model) Idenitfy(ar1, ar2 AnnotationResults) []ThreeDimensionalPlots {
	const MinSize int = 40
	const MaxLoss float64 = 30
	const MinConf float32 = 0.3
	ret := make([]ThreeDimensionalPlots, 0)
	for i, tj1 := range ar1.Trajectories {
		for j, tj2 := range ar2.Trajectories {
			if tj1.conf*tj2.conf < MinConf {
				continue
			}
			pl, err := NewSyncedPlots(tj1, tj2)
			if err != nil {
				continue
			}
			ps, loss := m.Convert(pl)
			size := ps.RawMatrix().Rows
			if size < MinSize || loss > MaxLoss {
				continue
			}
			ret = append(ret, ThreeDimensionalPlots{
				i: i, j: j,
				loss:  loss,
				size:  size,
				plots: ps,
				start: pl.start,
				end:   pl.end,
			})
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].loss < ret[j].loss })
	return ret
}

func (m Model) Convert(plots *SyncedPlots) (*mat.Dense, float64) {
	m1, m2 := m.project(m.params, plots)
	ret := mat.NewDense(plots.size, 3, nil)
	ret.Add(m1, m2)
	ret.Scale(0.5, ret)

	diff := mat.NewDense(plots.size, 3, nil)
	diff.Sub(m1, m2)
	loss := .0
	for i := 0; i < diff.RawMatrix().Rows; i++ {
		loss += mat.NewVecDense(3, diff.RawRowView(i)).Norm(2)
	}
	loss /= float64(plots.size)

	return ret, loss
}
