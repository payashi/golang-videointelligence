package vannotate

import (
	"fmt"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type ScreenPlot struct {
	P, Q float64
}

type Series struct {
	Conf       float32
	Start, End int
	Plots      []ScreenPlot
}

func (sr Series) Len() float64 {
	ret := .0
	for i := sr.Start; i < sr.End; i++ {
		cp, cq := sr.Plots[i].P, sr.Plots[i].Q
		np, nq := sr.Plots[i+1].P, sr.Plots[i+1].Q
		dist := math.Sqrt((np-cp)*(np-cp) + (nq-cq)*(nq-cq))
		ret += dist
	}
	return ret
}

func PlotScreen(outDir, fileName string, srList []Series) {
	const minConf float32 = 0.2
	const ratio float64 = 16. / 9. // aspect ratio
	p := plot.New()

	p.X.Min = -0.5
	p.Y.Min = -0.5 / ratio
	p.X.Max = +0.5
	p.Y.Max = +0.5 / ratio

	p.Add(plotter.NewGrid())

	for i, sr := range srList {
		if sr.Conf < minConf {
			continue
		}
		plots := make(plotter.XYs, len(sr.Plots))
		for i, v := range sr.Plots {
			plots[i].X = v.P
			plots[i].Y = v.Q
		}
		ploti, err := plotter.NewScatter(plots)
		if err != nil {
			panic(err)
		}

		ploti.GlyphStyle.Color = plotutil.Color(i)
		ploti.GlyphStyle.Radius = 2
		p.Add(ploti)

		p.Legend.Add(fmt.Sprintf("tr-%2d [%03d-%03d]", i, sr.Start, sr.End), ploti)
	}
	pwidth := 6 * vg.Inch
	pheight, _ := vg.ParseLength(fmt.Sprintf("%.2fin", 6/ratio))

	if err := p.Save(pwidth, pheight, fmt.Sprintf("%s/%s.png", outDir, fileName)); err != nil {
		panic(err)
	}
}
