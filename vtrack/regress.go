package vtrack

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

func Regress(tj1 Trajectory, tj2 Trajectory) {
	start := MaxInt(tj1.Start, tj2.Start)
	end := MinInt(tj1.End, tj2.End)
	if start > end {
		log.Fatal("These Trajectories have no common segments.")
	}
	tr1 := tj1.Plots[start : end+1]
	tr2 := tj2.Plots[start : end+1]
	model := new(Model)
	model.m = int(end - start + 1) // the number of datasets
	model.data = make([]Snapshot, model.m)
	for i := 0; i < model.m; i++ {
		p1, q1 := tr1.XY(i)
		p2, q2 := tr2.XY(i)
		model.data[i] = Snapshot{
			p1 - tj1.Width/2,
			p2 - tj2.Height/2,
			q1 - tj1.Width/2,
			q2 - tj2.Height/2,
		}
		// fmt.Printf("%.2f, %.2f, %.2f, %.2f\n", p1, q1, p2, q2)
	}
	model.c1 = mat.NewVecDense(3, []float64{0, 0, 0})
	model.c2 = mat.NewVecDense(3, []float64{0, 1, 0})
	model.theta1, model.theta2 = 0, 0
	model.phi1 = rand.Float64() * math.Pi
	model.phi2 = -rand.Float64() * math.Pi
	model.k1, model.k2 = 1, 1
	fmt.Printf("loss: %.3f\n", model.GetLoss())
	model.theta1 += math.Pi * 0.5
	fmt.Printf("loss: %.3f\n", model.GetLoss())
}

type Snapshot struct {
	p1, p2, q1, q2 float64
}

type Model struct {
	theta1, theta2 float64
	phi1, phi2     float64
	k1, k2         float64
	m              int
	data           []Snapshot
	c1, c2         *mat.VecDense
}

func (model *Model) GetLoss() float64 {

	n1 := mat.NewVecDense(3, []float64{
		math.Cos(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.theta1),
	})
	n2 := mat.NewVecDense(3, []float64{
		math.Cos(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.theta1),
	})
	a1 := mat.NewVecDense(3, []float64{
		-math.Sin(model.phi1),
		-math.Cos(model.phi1),
		0,
	})
	a2 := mat.NewVecDense(3, []float64{
		-math.Sin(model.phi2),
		-math.Cos(model.phi2),
		0,
	})
	b1 := mat.NewVecDense(3, []float64{
		-math.Cos(model.phi1) * math.Sin(model.theta1),
		-math.Sin(model.phi1) * math.Sin(model.theta1),
		math.Cos(model.theta1),
	})
	b2 := mat.NewVecDense(3, []float64{
		-math.Cos(model.phi2) * math.Sin(model.theta2),
		-math.Sin(model.phi2) * math.Sin(model.theta2),
		math.Cos(model.theta2),
	})

	loss := .0

	for i := 0; i < model.m; i++ {
		snap := model.data[i]
		d1 := mat.NewVecDense(3, nil)
		d1.AddVec(ScaledVec(snap.p1, a1), ScaledVec(snap.q1, b1))
		d1.ScaleVec(model.k1, d1)
		d1.AddVec(n1, d1)

		d2 := mat.NewVecDense(3, nil)
		d2.AddVec(ScaledVec(snap.p2, a2), ScaledVec(snap.q2, b2))
		d2.ScaleVec(model.k2, d2)
		d2.AddVec(n2, d2)

		mata := mat.NewDense(2, 2, []float64{
			2 * mat.Dot(d1, d1), -2 * mat.Dot(d1, d2),
			-2 * mat.Dot(d1, d2), 2 * mat.Dot(d2, d2),
		})
		c12 := mat.NewVecDense(3, nil)
		c12.SubVec(model.c1, model.c2)
		matb := mat.NewVecDense(2, []float64{
			-mat.Dot(c12, d1),
			mat.Dot(c12, d2),
		})
		t := mat.NewVecDense(2, nil)
		matainv := mat.NewDense(2, 2, nil)
		error := matainv.Inverse(mata)
		if error != nil {
			log.Fatal("Inversed matrix does not exist")
		}
		t.MulVec(matainv, matb)
		mate1 := ScaledVec(t.At(0, 0), d1)
		mate1.AddVec(model.c1, mate1)
		mate2 := ScaledVec(t.At(1, 0), d2)
		mate2.AddVec(model.c2, mate2)
		mate1.SubVec(mate1, mate2)
		loss += mat.Dot(mate1, mate1)
	}
	return loss
}

func ScaledVec(alpha float64, vec *mat.VecDense) *mat.VecDense {
	ret := mat.NewVecDense(vec.Len(), nil)
	ret.ScaleVec(alpha, vec)
	return ret
}

func MaxInt(nums ...int64) int64 {
	var ret int64 = math.MinInt64
	for _, v := range nums {
		if ret < v {
			ret = v
		}
	}
	return ret
}

func MinInt(nums ...int64) int64 {
	var ret int64 = math.MaxInt64
	for _, v := range nums {
		if ret > v {
			ret = v
		}
	}
	return ret
}
