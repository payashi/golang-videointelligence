package vtrack

import (
	"fmt"
	"log"
	"math"
	"sort"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type ThreeDimensionalPlots struct {
	loss  float64
	size  int
	plots *mat.Dense
	i, j  int
}

func Plot(outDir, fileName string, tdps []ThreeDimensionalPlots) {
	p := plot.New()
	for _, tdp := range tdps {
		pts := make(plotter.XYs, tdp.size)
		for i := 0; i < tdp.size; i++ {
			pts[i] = plotter.XY{X: tdp.plots.At(i, 0), Y: tdp.plots.At(i, 1)}
		}
		err := plotutil.AddLinePoints(p, pts)
		if err != nil {
			panic(err)
		}
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
	const MinSize int = 10
	const MaxLoss float64 = 0.1
	ret := make([]ThreeDimensionalPlots, 0)
	for i, tj1 := range ar1.Trajectories {
		for j, tj2 := range ar2.Trajectories {
			if i != 0 || j != 10 {
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
			})
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].loss < ret[j].loss })
	return ret
}

func (m Model) Convert(plots *SyncedPlots) (*mat.Dense, float64) {
	theta1, theta2 := m.params.At(0, 0), m.params.At(1, 0)

	n1 := mat.NewVecDense(3, []float64{
		math.Cos(m.config.Phi1) * math.Cos(theta1),
		math.Sin(m.config.Phi1) * math.Cos(theta1),
		math.Sin(theta1),
	})
	n2 := mat.NewVecDense(3, []float64{
		math.Cos(m.config.Phi2) * math.Cos(theta2),
		math.Sin(m.config.Phi2) * math.Cos(theta2),
		math.Sin(theta2),
	})
	a1 := mat.NewVecDense(3, []float64{
		math.Sin(m.config.Phi1),
		-math.Cos(m.config.Phi1),
		0,
	})
	a2 := mat.NewVecDense(3, []float64{
		math.Sin(m.config.Phi2),
		-math.Cos(m.config.Phi2),
		0,
	})
	b1 := mat.NewVecDense(3, []float64{
		-math.Cos(m.config.Phi1) * math.Sin(theta1),
		-math.Sin(m.config.Phi1) * math.Sin(theta1),
		math.Cos(theta1),
	})
	b2 := mat.NewVecDense(3, []float64{
		-math.Cos(m.config.Phi2) * math.Sin(theta2),
		-math.Sin(m.config.Phi2) * math.Sin(theta2),
		math.Cos(theta2),
	})

	c21 := mat.NewVecDense(3, nil)
	c21.SubVec(&m.config.C2, &m.config.C1)

	ret := mat.NewDense(plots.size, 3, nil)
	loss := .0
	for i := 0; i < plots.size; i++ {
		pl1 := plots.pl1[i]
		pl2 := plots.pl2[i]

		d1 := mat.NewVecDense(3, nil)
		d1.AddScaledVec(d1, pl1.p, a1)
		d1.AddScaledVec(d1, pl1.q, b1)
		d1.AddScaledVec(n1, m.config.K1, d1)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, pl2.p, a2)
		d2.AddScaledVec(d2, pl2.q, b2)
		d2.AddScaledVec(n2, m.config.K2, d2)

		mata := mat.NewDense(2, 2, []float64{
			mat.Dot(d1, d1), -mat.Dot(d1, d2),
			mat.Dot(d1, d2), -mat.Dot(d2, d2),
		})
		matb := mat.NewVecDense(2, []float64{
			mat.Dot(c21, d1),
			mat.Dot(c21, d2),
		})
		t := mat.NewVecDense(2, nil)
		matainv := mat.NewDense(2, 2, nil)
		error := matainv.Inverse(mata)
		if error != nil {
			log.Fatal("Inversed matrix does not exist")
		}
		t.MulVec(matainv, matb)
		l1 := mat.NewVecDense(3, nil)
		l1.AddScaledVec(&m.config.C1, t.At(0, 0), d1)
		l2 := mat.NewVecDense(3, nil)
		l2.AddScaledVec(&m.config.C2, t.At(1, 0), d2)
		// Mid Vector
		lm := mat.NewVecDense(3, nil)
		lm.AddVec(l1, l2)
		lm.ScaleVec(0.5, lm)
		ret.SetRow(i, lm.RawVector().Data)

		// Diff Vector
		ld := mat.NewVecDense(3, nil)
		ld.SubVec(l1, l2)
		loss += ld.Norm(2)
	}
	loss /= float64(plots.size)
	return ret, loss
}
