package vtrack

import (
	"errors"
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
			ip, err := cs.newIplots(sr1, sr2)
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

func (cs CameraSystem) newIplots(sr1, sr2 vannotate.Series) (IPlots, error) {
	m1 := cs.project(cs.params, 0, sr1.Plots)
	m2 := cs.project(cs.params, 1, sr2.Plots)
	// overwrapped range
	start, end := maxInt(sr1.Start, sr2.Start), minInt(sr1.End, sr2.End)
	if start > end {
		return IPlots{}, errors.New("no overwrapped range")
	}

	ret := IPlots{}
	ret.sr1, ret.sr2 = sr1, sr2
	ret.Start = minInt(sr1.Start, sr2.Start)
	ret.End = maxInt(sr1.End, sr2.End)
	ret.Size = ret.End - ret.Start + 1

	// Calculate loss
	ret.Loss = .0
	for t := start; t <= end; t++ {
		diff := mat.NewVecDense(3, nil)
		diff.SubVec(m1.RowView(t), m2.RowView(t))
		ret.Loss += diff.Norm(2)
	}
	ret.Loss /= float64(end - start + 1)

	ret.Plots = mat.NewDense(ret.Size, 3, nil)
	for t := ret.Start; t <= ret.End; t++ {
		// p1, p2 := sr1.Plots[t], sr2.Plots[t]
		in1 := sr1.Start <= t && t <= sr1.End
		in2 := sr2.Start <= t && t <= sr2.End
		p := mat.NewVecDense(3, nil)
		if in1 && in2 {
			p.AddVec(m1.RowView(t), m2.RowView(t))
			p.ScaleVec(0.5, p)
			ret.Plots.SetRow(t-ret.Start, p.RawVector().Data)
		} else if in1 {
			ret.Plots.SetRow(t-ret.Start, m1.RawRowView(t))
		} else if in2 {
			ret.Plots.SetRow(t-ret.Start, m2.RawRowView(t))
		} else {
			panic("invalid timestamp")
		}
	}

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
