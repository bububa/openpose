package openpose

import (
	"errors"
	"image"
	"io/ioutil"
	"math"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/bububa/openpose/gaussian"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// PoseEstimator represents pose estimator instance
type PoseEstimator struct {
	model        *tf.SavedModel
	modelPath    string
	modelName    string
	modelTags    []string
	modelType    ModelType
	sharpenSigma float64
	minSize      float64
	scaleFactor  float64
	scaleTimes   int
	upsampleSize int
	mutex        sync.Mutex
}

// NewPoseEstimator returns a new TensorFlow PoseEstimator instance.
func NewPoseEstimator(modelPath string, modelType ModelType) *PoseEstimator {
	return &PoseEstimator{
		modelPath:    modelPath,
		modelType:    modelType,
		modelTags:    []string{"serve"},
		minSize:      5,
		scaleFactor:  0.709,
		scaleTimes:   5,
		upsampleSize: 4,
		sharpenSigma: DefaultSharpenSigma,
	}
}

// SetSharpenSigma set sharpen sigma for image preprocessing
func (t *PoseEstimator) SetSharpenSigma(sigma float64) {
	t.sharpenSigma = sigma
}

// Estimate returns estimated Humans in an image
func (t *PoseEstimator) Estimate(img image.Image, modelSize ModelSize) ([]Human, error) {
	if err := t.LoadModel(); err != nil {
		return nil, err
	}
	/*
		for _, operation := range t.model.Graph.Operations() {
			log.Println(operation.Name())
		}
	*/
	preprocessedImage, normPadding := ImagePreprocess(img, modelSize, t.sharpenSigma)
	return t.estimatePose(preprocessedImage, normPadding)
}

func (t *PoseEstimator) estimatePose(img image.Image, normPadding Size) ([]Human, error) {
	pafMat, heatMat, err := t.getMats(img)
	if err != nil {
		return nil, err
	}
	heatMat = gaussian.ApplyFilter(heatMat, 5, 2.5)
	//dump(pafMat, "./pafMat.json")
	//dump(heatMat, "./heatMat.json")
	nmsThreshold := math.Max(float64(matAverage(heatMat)*4), NMS_Threshold)
	nmsThreshold = math.Min(nmsThreshold, 0.3)
	//log.Printf("nms, th=%f, mat:%f\n", nmsThreshold, matAverage(heatMat))
	scales := scales(float64(len(heatMat[0])), float64(len(heatMat[0][0])), t.scaleFactor, t.minSize)
	coords := make([][2][]int, 0, TotalBodyParts)
	nmsThresholdf32 := float32(nmsThreshold)
	for _, plain := range heatMat[0:TotalBodyParts] {
		nms, err := nonMaxSuppression(plain, scales[0], nmsThresholdf32)
		if err != nil {
			return nil, err
		}
		coords = append(coords, maxiumFilter(nms, nmsThresholdf32))
	}
	//dump(coords, "./coords.json")

	// connect parts
	var connections []Connection
	connectionPool := &sync.Pool{
		New: func() interface{} {
			return new(Connection)
		},
	}
	for idx, cocoPair := range CocoPairs {
		pairNetwork := CocoPairsNetwork[idx]
		conns := t.estimatePosePair(connectionPool, coords, cocoPair[0], cocoPair[1], pafMat[pairNetwork[0]], pafMat[pairNetwork[1]], heatMat, normPadding)
		connections = append(connections, conns...)
	}

	heatMatRows := float64(len(heatMat[0]))
	heatMatCols := float64(len(heatMat[0][0]))
	return connectionsToHumans(connections, heatMatRows, heatMatCols), nil
}

func (t *PoseEstimator) estimatePosePair(connectionPool *sync.Pool, coords [][2][]int, part1 CocoPart, part2 CocoPart, pafMatX [][]float32, pafMatY [][]float32, heatMat [][][]float32, normPadding Size) []Connection {
	peakCoord1, peakCoord2 := coords[part1], coords[part2]
	var abovePairs = [][2]CocoPart{
		{CocoPartRShoulder}, {CocoPartRElbow},
		{CocoPartRElbow}, {CocoPartRWrist},
		{CocoPartLShoulder}, {CocoPartLElbow},
		{CocoPartLElbow}, {CocoPartLWrist},
	}
	var inAboveParts bool
	for _, pair := range abovePairs {
		if part1 == pair[0] && part1 == pair[1] {
			inAboveParts = true
			break
		}
	}
	candidates := make([]Connection, 0, len(peakCoord1[0])*len(peakCoord2[0]))
	for idx1, y1 := range peakCoord1[0] {
		x1 := peakCoord1[1][idx1]
		for idx2, y2 := range peakCoord2[0] {
			x2 := peakCoord2[1][idx2]
			x1f64, y1f64, x2f64, y2f64 := float64(x1), float64(y1), float64(x2), float64(y2)
			score, count := t.getScore(x1f64, y1f64, x2f64, y2f64, pafMatX, pafMatY)
			//log.Printf("part:%d-%d, score:%f, count:%d, p1:%d-%d, p2:%d-%d\n", part1, part2, score, count, x1, y1, x2, y2)
			if inAboveParts && count < InterMinAboveThreshold {
				continue
			} else if !inAboveParts && (count < InterMinAboveThreshold || score <= 0) {
				continue
			}
			candidate := connectionPool.Get().(*Connection)
			candidate.NormPadding = normPadding
			candidate.Score = score
			candidate.Coords = [2]image.Point{
				image.Pt(x1, y1),
				image.Pt(x2, y2),
			}
			candidate.Idx = [2]int{idx1, idx2}
			candidate.Parts = [2]CocoPart{part1, part2}
			candidate.Scores = [2]float32{
				heatMat[part1][y1][x1],
				heatMat[part2][y2][x2],
			}
			candidate.UPartIdx = candidate.GetUPartIdx()
			connectionPool.Put(candidate)
			candidates = append(candidates, *candidate)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
	connections := make([]Connection, 0, len(candidates))
	var (
		usedIdx1 = make(map[int]struct{}, len(peakCoord1[0]))
		usedIdx2 = make(map[int]struct{}, len(peakCoord2[0]))
	)
	for _, candidate := range candidates {
		// check not connected
		_, found1 := usedIdx1[candidate.Idx[0]]
		_, found2 := usedIdx2[candidate.Idx[1]]
		if found1 || found2 {
			continue
		}
		connections = append(connections, candidate)
		usedIdx1[candidate.Idx[0]] = struct{}{}
		usedIdx2[candidate.Idx[1]] = struct{}{}
	}
	return connections
}

func (t *PoseEstimator) getScore(x1, y1, x2, y2 float64, pafMatX, pafMatY [][]float32) (float32, int) {
	dx, dy := x2-x1, y2-y1
	normVec := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))
	if normVec < 1e-4 {
		return 0, 0
	}
	vx, vy := float32(dx/normVec), float32(dy/normVec)
	var (
		numIter  int     = 10
		numIterf float64 = 10
	)

	stepX, stepY := dx/numIterf, dy/numIterf
	var (
		xs = make([]int, 0, numIter)
		ys = make([]int, 0, numIter)
	)
	var i int
	xv, yv := x1, y1
	for i < numIter {
		xs = append(xs, int(xv+0.5))
		ys = append(ys, int(yv+0.5))
		if x1 != x2 {
			xv += stepX
		}
		if y1 != y2 {
			yv += stepY
		}
		i++
	}
	var (
		score float32
		count int
	)
	for idx, x := range xs {
		y := ys[idx]
		pafX := pafMatX[y][x]
		pafY := pafMatY[y][x]
		localScore := pafX*vx + pafY*vy
		if localScore > InterThreashold {
			score += localScore
			count++
		}
	}
	return score, count
}

func (t *PoseEstimator) getMats(img image.Image) ([][][]float32, [][][]float32, error) {
	tensor, err := makeTensorFromImage(img)
	if err != nil {
		return nil, nil, err
	}
	var (
		inOp  map[tf.Output]*tf.Tensor
		outOp []tf.Output
	)

	switch t.modelType {
	case CMU:
		inOp = map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("image").Output(0): tensor,
		}
		outOp = []tf.Output{
			t.model.Graph.Operation("Mconv7_stage6_L1/BiasAdd").Output(0),
			t.model.Graph.Operation("Mconv7_stage6_L2/BiasAdd").Output(0),
		}
	case MobileNet:
		inOp = map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("image").Output(0): tensor,
		}
		outOp = []tf.Output{
			t.model.Graph.Operation("Openpose/MConv_Stage6_L1_5_pointwise/BatchNorm/FusedBatchNorm").Output(0),
			t.model.Graph.Operation("Openpose/MConv_Stage6_L2_5_pointwise/BatchNorm/FusedBatchNorm").Output(0),
		}
	default:
		return nil, nil, errors.New("invalid model type")
	}

	output, err := t.model.Session.Run(
		inOp,
		outOp,
		nil)

	if err != nil {
		return nil, nil, err
	}
	if len(output) != 2 {
		return nil, nil, errors.New("inference failed, no output")
	}
	pafMat := output[0].Value().([][][][]float32)[0]
	heatMat := output[1].Value().([][][][]float32)[0]
	if output[1].Shape()[3] == 19 {
		heatMat = rollAxis(heatMat, 2, 0)
	}
	if output[0].Shape()[3] == 38 {
		pafMat = rollAxis(pafMat, 2, 0)
	}
	return pafMat, heatMat, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *PoseEstimator) ModelLoaded() bool {
	return t.model != nil
}

// LoadModel load tensorfow model
func (t *PoseEstimator) LoadModel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelPath)

	// Load model
	//model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)
	fn, err := os.Open(modelPath)
	if err != nil {
		return err
	}
	defer fn.Close()
	graph := tf.NewGraph()
	data, err := ioutil.ReadAll(fn)
	if err != nil {
		return err
	}
	if err := graph.Import(data, ""); err != nil {
		return err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return err
	}

	model := &tf.SavedModel{
		Graph:   graph,
		Session: session,
	}
	t.model = model

	return nil
}
