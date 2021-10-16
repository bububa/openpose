package gaussian

import "math"

func getGaussianKernel(size int, sigma float64) ([][]float32, float32) {
	// https://homepages.inf.ed.ac.uk/rbf/HIPR2/gsmooth.htm
	kSize := size

	kern1d := make([]float64, kSize)
	kern2d := make([][]float64, kSize)
	gaussianFilter := make([][]float32, kSize)

	// initialize matrices
	for i := range kern2d {
		kern2d[i] = make([]float64, kSize)
		gaussianFilter[i] = make([]float32, kSize)
	}

	// Calculate 1-D Gaussian distribution
	twoSigmaSq := 2 * math.Pow(sigma, 2)
	calc1 := 1.0 / (math.Sqrt(2*math.Pi) * sigma)

	for i := -size / 2; i < (size/2)+1; i++ {
		numerator := math.Pow(float64(i), 2)
		kern1d[i+(size/2)] = calc1 * math.Exp(-(numerator / twoSigmaSq))
	}

	//outer product
	kern2dCsum := 0.0
	for i := range kern1d {
		for j := range kern1d {
			mult := kern1d[i] * kern1d[j]
			kern2d[i][j] = mult
			kern2dCsum += mult
		}
	}

	// normalize
	scalar := 1.0 / kern2d[0][0]
	for i := range kern1d {
		for j := range kern1d {
			gaussianFilter[i][j] = float32(math.Floor((kern2d[i][j] / kern2dCsum) * scalar))
		}
	}

	return gaussianFilter, float32(scalar)
}

// ApplyFilter returns gaussian filtered array
func ApplyFilter(data [][][]float32, window int, sigma float64) [][][]float32 {
	rows := len(data[0])
	cols := len(data[0][0])
	nKer, kScalar := getGaussianKernel(window, sigma)
	var kUpper int = (len(nKer) / 2) + 1
	var kLower int = len(nKer) / 2
	ret := make([][][]float32, len(data))
	for partId, part := range data {
		ret[partId] = make([][]float32, rows)
		for i, _ := range data[0] {
			ret[partId][i] = make([]float32, cols)
		}
		// Convolve filter mask over image
		for y := kLower; y < len(part)-kLower; y++ {
			for x := kLower; x < len(part[y])-kLower; x++ {
				var sum float32
				// iterate over kernel
				for i := -kLower; i < kUpper; i++ {
					for j := -kLower; j < kUpper; j++ {
						val := part[y+i][x+j]
						sum += val * nKer[i+kLower][j+kLower]
					}
				}
				// calculate sum average
				ret[partId][y][x] = sum / kScalar
			}
		}
	}

	return ret
}
