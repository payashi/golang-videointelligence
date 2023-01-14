package vtrack

import (
	"fmt"
	"math"
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
	n1, n2 := len(ar1.Trajectories), len(ar2.Trajectories)
	const MinSize int = 40
	const MaxLoss float64 = 30
	const MinDist float64 = 10.
	tdps := make([][]ThreeDimensionalPlots, n1)
	for i := 0; i < n1; i++ {
		tdps[i] = make([]ThreeDimensionalPlots, n2)
		for j := 0; j < n2; j++ {
			tdps[i][j] = ThreeDimensionalPlots{
				i: -1, j: -1,
				loss:  math.Inf(1),
				size:  0,
				plots: &mat.Dense{},
				start: 0,
				end:   0,
			}

		}
	}
	for i, tj1 := range ar1.Trajectories {
		for j, tj2 := range ar2.Trajectories {
			pl, err := NewSyncedPlots(tj1, tj2)
			if err != nil {
				continue
			}
			ps, loss := m.Convert(pl)
			size := ps.RawMatrix().Rows
			if size < MinSize || loss > MaxLoss {
				continue
			}
			path := mat.NewVecDense(3, nil)
			path.SubVec(ps.RowView(size-1), ps.RowView(0))
			if path.Norm(2) < MinDist {
				continue
			}
			tdps[i][j] = ThreeDimensionalPlots{
				i: i, j: j,
				loss:  loss,
				size:  size,
				plots: ps,
				start: pl.start,
				end:   pl.end,
			}
		}
	}
	ret := make([]ThreeDimensionalPlots, 0)
	usedis := make([]int, 0)
	usedjs := make([]int, 0)
	for {
		argmini, argminj := -1, -1
		for i := 0; i < n1; i++ {
			if contains(usedis, i) {
				continue
			}
			best := math.Inf(1)
			for j := 0; j < n2; j++ {
				if contains(usedjs, j) {
					continue
				}
				tdp := &tdps[i][j]
				if best > tdp.loss {
					best = tdp.loss
					argmini, argminj = tdp.i, tdp.j
				}
			}
		}
		if argmini == -1 {
			break
		}
		usedis = append(usedis, argmini)
		usedjs = append(usedjs, argminj)
		ret = append(ret, tdps[argmini][argminj])
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
	for i := 0; i < plots.size; i++ {
		loss += mat.NewVecDense(3, diff.RawRowView(i)).Norm(2)
	}
	loss /= float64(plots.size)

	return ret, loss
}

func contains(s []int, t int) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}
	return false
}
