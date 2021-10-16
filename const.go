package openpose

const TotalBodyParts = 18

// CocoPart represents body parts
type CocoPart int

const (
	// CocoPartNose nose
	CocoPartNose CocoPart = iota
	// CocoPartNeck neck
	CocoPartNeck
	// CocoPartRShoulder right sholder
	CocoPartRShoulder
	// CocoPartRElbow right elbow
	CocoPartRElbow
	// CocoPartRWrist right wrist
	CocoPartRWrist
	// CocoPartLShoulder left sholder
	CocoPartLShoulder
	// CocoPartLElbow left elbow
	CocoPartLElbow
	// CocoPartLWrist left wrist
	CocoPartLWrist
	// CocoPartRHip right hip
	CocoPartRHip
	// CocoPartRKnee right knee
	CocoPartRKnee
	// CocoPartRAnkle right ankle
	CocoPartRAnkle
	// CocoPartLHip left hip
	CocoPartLHip
	// CocoPartLKnee left knee
	CocoPartLKnee
	// CocoPartLAnkle left ankle
	CocoPartLAnkle
	// CocoPartREye right eye
	CocoPartREye
	// CocoPartLEye left eye
	CocoPartLEye
	// CocoPartREar right ear
	CocoPartREar
	// CocoPartLEar left ear
	CocoPartLEar
	// CocoPartBackground background
	CocoPartBackground
)

// MPIIPart MPII human parts
type MPIIPart int

const (
	// MPIIPartRAnkle right ankle
	MPIIPartRAnkle MPIIPart = iota
	// MPIIPartRKnee right knee
	MPIIPartRKnee

	MPIIPartRHip
	// MPIIPartLHip left hip
	MPIIPartLHip
	// MPIIPartLKnee left knee
	MPIIPartLKnee
	// MPIIPartLAnkle left ankle
	MPIIPartLAnkle
	// MPIIPartRWrist right wrist
	MPIIPartRWrist
	// MPIIPartRElbow right elbow
	MPIIPartRElbow
	// MPIIPartRShoulder right sholder
	MPIIPartRShoulder
	// MPIIPartLShoulder left sholder
	MPIIPartLShoulder
	// MPIIPartLElbow left elbow
	MPIIPartLElbow
	// MPIIPartLWrist left wrist
	MPIIPartLWrist
	// MPIIPartNeck neck
	MPIIPartNeck
	// MPIIPartHead head
	MPIIPartHead
)

// PartPair represents pose part MPIIPart, CocoPart pair
type PartPair struct {
	MPIIPart MPIIPart
	CocoPart CocoPart
}

// PartPairs represents MPIIPart, CocoPart pair list
var PartPairs = []PartPair{
	{MPIIPartHead, CocoPartNose},
	{MPIIPartNeck, CocoPartNeck},
	{MPIIPartRShoulder, CocoPartRShoulder},
	{MPIIPartRElbow, CocoPartRElbow},
	{MPIIPartRWrist, CocoPartRWrist},
	{MPIIPartLShoulder, CocoPartLShoulder},
	{MPIIPartLElbow, CocoPartLElbow},
	{MPIIPartLWrist, CocoPartLWrist},
	{MPIIPartRHip, CocoPartRHip},
	{MPIIPartRKnee, CocoPartRKnee},
	{MPIIPartRAnkle, CocoPartRAnkle},
	{MPIIPartLHip, CocoPartLHip},
	{MPIIPartLKnee, CocoPartLKnee},
	{MPIIPartLAnkle, CocoPartLAnkle},
}

// CocoPairs coco part pairs
var CocoPairs = [][2]CocoPart{
	{CocoPartNeck, CocoPartRShoulder}, {CocoPartNeck, CocoPartLShoulder}, {CocoPartRShoulder, CocoPartRElbow},
	{CocoPartRElbow, CocoPartRWrist}, {CocoPartLShoulder, CocoPartLElbow}, {CocoPartLElbow, CocoPartLWrist},
	{CocoPartNeck, CocoPartRHip}, {CocoPartRHip, CocoPartRKnee}, {CocoPartRKnee, CocoPartRAnkle},
	{CocoPartNeck, CocoPartLHip}, {CocoPartLHip, CocoPartLKnee}, {CocoPartLKnee, CocoPartLAnkle},
	{CocoPartNeck, CocoPartNose},
	{CocoPartNose, CocoPartREye}, {CocoPartREye, CocoPartREar},
	{CocoPartNose, CocoPartLEye}, {CocoPartLEye, CocoPartLEar},
	{CocoPartRShoulder, CocoPartREar},
	{CocoPartLShoulder, CocoPartLEar},
}

// CocoPairNetwork .
var CocoPairsNetwork = [][2]CocoPart{
	{12, 13}, {20, 21}, {14, 15}, {16, 17}, {22, 23}, {24, 25}, {0, 1}, {2, 3}, {4, 5}, {6, 7}, {8, 9}, {10, 11}, {28, 29}, {30, 31}, {34, 35}, {32, 33}, {36, 37}, {18, 19}, {26, 27},
}

// CocoPairsRender represents coco pairs for render
var CocoPairsRender = CocoPairs[0 : len(CocoPairs)-2]

// CocoColors represents color for coco parts
var CocoColors = [][3]uint8{
	{255, 0, 0}, {255, 85, 0}, {255, 170, 0}, {255, 255, 0}, {170, 255, 0}, {85, 255, 0}, {0, 255, 0}, {0, 255, 85}, {0, 255, 170}, {0, 255, 255}, {0, 170, 255}, {0, 85, 255}, {0, 0, 255}, {85, 0, 255}, {170, 0, 255}, {255, 0, 255}, {255, 0, 170}, {255, 0, 85},
}

// CoordParts represents CocoParts for coordinate
var CoordParts = []CocoPart{
	CocoPartNose,
	CocoPartNeck,
	CocoPartRShoulder,
	CocoPartLShoulder,
	CocoPartRHip,
	CocoPartLHip,
	CocoPartREye,
	CocoPartLEye,
	CocoPartREar,
	CocoPartLEar,
}

// CoordPartsMap represents CocoParts for coordinate in map
var CoordPartsMap = map[CocoPart]struct{}{
	CocoPartNose:      {},
	CocoPartNeck:      {},
	CocoPartRShoulder: {},
	CocoPartLShoulder: {},
	CocoPartRHip:      {},
	CocoPartLHip:      {},
	CocoPartREye:      {},
	CocoPartLEye:      {},
	CocoPartREar:      {},
	CocoPartLEar:      {},
}

const (
	ThresholdPartConfidence float32 = 0.3
	InterThreashold         float32 = 0.1
	InterMinAboveThreshold  int     = 6
	NMS_Threshold           float64 = 0.1
	MinSubsetCnt            int     = 4
	MinSubsetScore          float32 = 0.8
	ThresholdHumanScore     float32 = 0.4
)

const (
	DefaultSharpenSigma float64 = 0.0
)

// ModelSize size for training image
type ModelSize [2]int

var (
	ModelSizeBest    ModelSize = [2]int{1312, 736}
	ModelSizeBetter  ModelSize = [2]int{656, 368}
	ModelSizeCMU     ModelSize = [2]int{640, 360}
	ModelSizeDefault ModelSize = [2]int{432, 368}
	ModelSizeFaster  ModelSize = [2]int{336, 288}
	ModelSizeFatest  ModelSize = [2]int{304, 240}
)

// ModelType represents type of mode graph
type ModelType int

const (
	// CMU cmu model graph
	CMU ModelType = iota
	// MobileNetmodelnet model graph
	MobileNet
)
