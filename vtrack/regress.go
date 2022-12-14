package vtrack

import (
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
}

func (model *Model) getLoss() {
	// x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// A := mat.NewDense(3, 4, x)
	n1 := mat.NewDense(3, 1, []float64{
		math.Cos(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.theta1),
	})
	n2 := mat.NewDense(3, 1, []float64{
		math.Cos(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.phi1) * math.Cos(model.theta1),
		math.Sin(model.theta1),
	})
	a1 := mat.NewDense(3, 1, []float64{
		-math.Sin(model.phi1),
		-math.Cos(model.phi1),
		0,
	})
	a2 := mat.NewDense(3, 1, []float64{
		-math.Sin(model.phi2),
		-math.Cos(model.phi2),
		0,
	})
	b1 := mat.NewDense(3, 1, []float64{
		-math.Cos(model.phi1) * math.Sin(model.theta1),
		-math.Sin(model.phi1) * math.Sin(model.theta1),
		math.Cos(model.theta1),
	})
	b2 := mat.NewDense(3, 1, []float64{
		-math.Cos(model.phi2) * math.Sin(model.theta2),
		-math.Sin(model.phi2) * math.Sin(model.theta2),
		math.Cos(model.theta2),
	})

	for i := 0; i < model.m; i++ {
		snap := model.data[i]
		d1 := n1 + model.k1*(snap.p1*a1+snap.q1*b1)
		d2 := n2 + model.k2*(snap.p2*a2+snap.q2*b2)

	}

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
