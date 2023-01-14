package vtrack

import (
	"errors"
	"math"

	"github.com/payashi/vannotate"
)

// Synchronized Plots
type splots struct {
	pl1, pl2   []vannotate.ScreenPlot
	start, end int
	size       int
}

func NewSyncedPlots(sr1, sr2 vannotate.Series) (*splots, error) {
	start := maxInt(sr1.Start, sr2.Start)
	end := minInt(sr1.End, sr2.End)
	// Return error when there's no overwrap
	if start > end {
		return nil, errors.New("syncedplots: no overwrap")
	}
	return &splots{
		size:  end - start + 1,
		start: start,
		end:   end,
		pl1:   sr1.Plots[start : end+1],
		pl2:   sr2.Plots[start : end+1],
	}, nil
}

func maxInt(nums ...int) int {
	ret := math.MinInt
	for _, v := range nums {
		if ret < v {
			ret = v
		}
	}
	return ret
}

func minInt(nums ...int) int {
	ret := math.MaxInt
	for _, v := range nums {
		if ret > v {
			ret = v
		}
	}
	return ret
}
