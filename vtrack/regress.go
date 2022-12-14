package vtrack

import (
	"fmt"
	"log"
	"math"

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
	model.ndata = int(end - start + 1) // the number of datasets
	model.data = make([]Snapshot, model.ndata)
	for i := 0; i < model.ndata; i++ {
		p1, q1 := tr1.XY(i)
		p2, q2 := tr2.XY(i)
		model.data[i] = Snapshot{
			p1 - tj1.Width/2,
			p2 - tj2.Width/2,
			q1 - tj1.Height/2,
			q2 - tj2.Height/2,
		}
	}
	model.c1 = mat.NewVecDense(3, []float64{0, 0, 0})
	model.c2 = mat.NewVecDense(3, []float64{0, 1, 0})
	model.params = mat.NewVecDense(6, []float64{
		0, 0, // theta
		// rand.Float64() * math.Pi, -rand.Float64() * math.Pi, // phi
		0, 0,
		1, 1, // k
	})
	model.nparams = model.params.Len()

	dp := 1e-2
	mu := 1.
	for i := 0; i < 100; i++ {
		// fmt.Printf("loss: %.3f\n", model.GetLoss(*model.params))

		inc := mat.NewVecDense(model.nparams, nil)
		for i := 0; i < model.nparams; i++ {
			inc.SetVec(i, -model.GetDiff(i, dp))
		}
		inc.ScaleVec(1/inc.Norm(2), inc)
		model.params.AddScaledVec(model.params, mu, inc)
	}
	fmt.Printf("final loss: %.3f\n", model.GetLoss(*model.params))
	fmt.Printf("theta1: %.2f pi, theta2: %.2f pi\n", model.params.At(0, 0)/math.Pi, model.params.At(1, 0)/math.Pi)
	fmt.Printf("phi1: %.2f pi, phi2: %.2f pi\n", model.params.At(2, 0)/math.Pi, model.params.At(3, 0)/math.Pi)
	fmt.Printf("k1: %.2f, k2: %.2f\n", model.params.At(4, 0), model.params.At(5, 0))
}

type Snapshot struct {
	p1, p2, q1, q2 float64
}

type Model struct {
	nparams, ndata int
	params         *mat.VecDense
	data           []Snapshot
	c1, c2         *mat.VecDense
}

func (model *Model) GetDiff(i int, dp float64) float64 {
	cv := model.GetLoss(*model.params)
	nparams := mat.NewVecDense(model.nparams, nil)
	nparams.SetVec(i, dp)
	nparams.AddVec(model.params, nparams)
	nv := model.GetLoss(*nparams)
	return (nv - cv) / dp
}

func (model *Model) GetLoss(params mat.VecDense) float64 {
	theta1, theta2 := params.At(0, 0), params.At(1, 0)
	phi1, phi2 := params.At(2, 0), params.At(3, 0)
	k1, k2 := params.At(4, 0), params.At(5, 0)

	n1 := mat.NewVecDense(3, []float64{
		math.Cos(phi1) * math.Cos(theta1),
		math.Sin(phi1) * math.Cos(theta1),
		math.Sin(theta1),
	})
	n2 := mat.NewVecDense(3, []float64{
		math.Cos(phi1) * math.Cos(theta1),
		math.Sin(phi1) * math.Cos(theta1),
		math.Sin(theta1),
	})
	a1 := mat.NewVecDense(3, []float64{
		-math.Sin(phi1),
		-math.Cos(phi1),
		0,
	})
	a2 := mat.NewVecDense(3, []float64{
		-math.Sin(phi2),
		-math.Cos(phi2),
		0,
	})
	b1 := mat.NewVecDense(3, []float64{
		-math.Cos(phi1) * math.Sin(theta1),
		-math.Sin(phi1) * math.Sin(theta1),
		math.Cos(theta1),
	})
	b2 := mat.NewVecDense(3, []float64{
		-math.Cos(phi2) * math.Sin(theta2),
		-math.Sin(phi2) * math.Sin(theta2),
		math.Cos(theta2),
	})

	loss := .0

	for i := 0; i < model.ndata; i++ {
		snap := model.data[i]
		d1 := mat.NewVecDense(3, nil)
		d1.AddScaledVec(d1, snap.p1, a1)
		d1.AddScaledVec(d1, snap.q1, b1)
		d1.AddScaledVec(n1, k1, d1)

		d2 := mat.NewVecDense(3, nil)
		d2.AddScaledVec(d2, snap.p2, a2)
		d2.AddScaledVec(d2, snap.q2, b2)
		d2.AddScaledVec(n2, k2, d2)

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
		l1 := mat.NewVecDense(3, nil)
		l1.AddScaledVec(model.c1, t.At(0, 0), d1)
		l2 := mat.NewVecDense(3, nil)
		l2.AddScaledVec(model.c2, t.At(1, 0), d2)
		l1.SubVec(l1, l2)
		loss += mat.Dot(l1, l1)
	}
	return loss
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
