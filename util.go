package openpose

import (
	"encoding/json"
	"math"
	"os"

	"gorgonia.org/tensor"
)

// RoundInt returns int round
func RoundInt(v float64) int {
	return int(math.Round(v))
}

// rollAxis returns np.rollaxis
func rollAxis(values [][][]float32, axis int, start int) [][][]float32 {
	d1 := len(values)
	d2 := len(values[0])
	d3 := len(values[0][0])
	flat := make([]float32, 0, d1*d2*d3)
	for _, v1 := range values {
		for _, v2 := range v1 {
			for _, v := range v2 {
				flat = append(flat, v)
			}
		}
	}
	d := tensor.New(tensor.WithShape(d1, d2, d3), tensor.WithBacking(flat))
	dout, _ := d.RollAxis(axis, start, true)
	shape := dout.Shape()
	ret := make([][][]float32, 0, shape[0])
	var i int
	for i < shape[0] {
		var y int
		yv := make([][]float32, 0, shape[1])
		for y < shape[1] {
			var x int
			xv := make([]float32, 0, shape[2])
			for x < shape[2] {
				v, _ := dout.At(i, y, x)
				xv = append(xv, v.(float32))
				x++
			}
			yv = append(yv, xv)
			y++
		}
		ret = append(ret, yv)
		i++
	}
	return ret
}

// matAverage returns np.average
func matAverage(mat [][][]float32) float32 {
	var (
		total float32
		count float32
	)
	for _, v1 := range mat {
		for _, v2 := range v1 {
			for _, v := range v2 {
				total += v
				count++
			}
		}
	}
	return total / count
}

func maxiumFilter(mat [][]float32, threshold float32) [2][]int {
	var ret [2][]int
	for y, rows := range mat {
		for x, v := range rows {
			if v <= threshold {
				continue
			}
			ret[0] = append(ret[0], y)
			ret[1] = append(ret[1], x)
		}
		if len(ret[0]) == 0 {
			ret[0] = make([]int, 0)
		}
		if len(ret[1]) == 0 {
			ret[1] = make([]int, 0)
		}
	}
	return ret
}

func scales(h, w float64, factor, minSize float64) []float64 {
	minl := h
	if minl > w {
		minl = w
	}

	m := 12.0 / minSize
	minl = minl * m

	var scales []float64
	for count := 0; minl > 12.0; {
		scales = append(scales, m*math.Pow(factor, float64(count)))
		minl = minl * factor
		count++
	}

	return scales
}

func dump(obj interface{}, filePath string) {
	fn, _ := os.Create(filePath)
	defer fn.Close()
	json.NewEncoder(fn).Encode(obj)
}
