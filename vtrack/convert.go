package vtrack

import (
	"math"
	"sort"

	"github.com/payashi/vannotate"
	"gonum.org/v1/gonum/mat"
)

func (cs CameraSystem) Idenitfy(srList1, srList2 []vannotate.Series) []IPlots {
	n1, n2 := len(srList1), len(srList2)
	const MaxLoss float64 = 30
	const MinDist float64 = 10.
	tdps := make([][]IPlots, n1)
	for i := 0; i < n1; i++ {
		tdps[i] = make([]IPlots, n2)
		for j := 0; j < n2; j++ {
			tdps[i][j] = IPlots{
				i: -1, j: -1,
				Loss:  math.Inf(1),
				Size:  0,
				Plots: &mat.Dense{},
				Start: 0,
				End:   0,
			}

		}
	}
	for i, sr1 := range srList1 {
		for j, sr2 := range srList2 {
			ip, err := cs.join(sr1, sr2)
			ip.i = i
			ip.j = j
			if err != nil {
				continue
			}
			if ip.Loss > MaxLoss {
				continue
			}
			path := mat.NewVecDense(3, nil)
			path.SubVec(ip.Plots.RowView(ip.Size-1), ip.Plots.RowView(0))
			if path.Norm(2) < MinDist {
				continue
			}
			tdps[i][j] = ip
		}
	}
	ret := make([]IPlots, 0)
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
				if best > tdp.Loss {
					best = tdp.Loss
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
	sort.Slice(ret, func(i, j int) bool { return ret[i].Loss < ret[j].Loss })
	return ret
}

func (cs CameraSystem) join(sr1, sr2 vannotate.Series) (IPlots, error) {
	ret := IPlots{}
	ret.sr1, ret.sr2 = sr1, sr2
	ret.Start = minInt(sr1.Start, sr2.Start)
	ret.End = maxInt(sr1.End, sr2.End)

	plots, err := NewSyncedPlots(sr1, sr2)
	if err != nil {
		return ret, err
	}

	m1 := cs.project(cs.params, 0, sr1.Plots)
	m2 := cs.project(cs.params, 1, sr2.Plots)

	ret.Plots.Add(m1, m2)
	ret.Plots.Scale(0.5, ret.Plots)

	diff := mat.NewDense(plots.size, 3, nil)
	diff.Sub(m1, m2)
	ret.Loss = .0
	for i := 0; i < plots.size; i++ {
		ret.Loss += mat.NewVecDense(3, diff.RawRowView(i)).Norm(2)
	}
	ret.Loss /= float64(plots.size)

	return ret, nil
}

func contains(s []int, t int) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}
	return false
}
