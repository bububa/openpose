package openpose

import (
	"math"
)

// Human represents human structure
type Human struct {
	Parts map[CocoPart]BodyPart
	Score float32
}

// NewHuman returns a new Human with given BodyPartPairs
func NewHuman() *Human {
	h := &Human{
		Parts: make(map[CocoPart]BodyPart, TotalBodyParts),
	}
	return h
}

// Reset reset human
func (h *Human) Reset() {
	h.Parts = map[CocoPart]BodyPart{}
	h.Score = 0
}

// PartCount returns total number of body parts
func (h Human) PartCount() int {
	return len(h.Parts)
}

// HasPart returns specific part exists in humnan or not
func (h Human) HasPart(part CocoPart) bool {
	_, found := h.Parts[part]
	return found
}

// GetMaxScore returns the max score of body parts
func (h Human) GetMaxScore() float32 {
	var score float32
	for _, part := range h.Parts {
		if score < part.Score {
			score = part.Score
		}
	}
	return score
}

// GetFaceBox returns upper body box compared to img size (w, h)
func (h Human) GetFaceBox(imgW float64, imgH float64, mode int) Rectangle {
	parts := make([]BodyPart, 0, TotalBodyParts)
	partsMap := make(map[CocoPart]BodyPart, TotalBodyParts)
	partPoints := make([]Point, 0, len(CoordPartsMap))
	var (
		x  float64
		y  float64
		x2 float64
		y2 float64
	)
	for _, part := range h.Parts {
		if part.Score <= ThresholdPartConfidence {
			continue
		}
		parts = append(parts, part)
		partsMap[part.Part] = part
		if _, found := CoordPartsMap[part.Part]; found {
			point := Pt(imgW*part.Point.X, imgH*part.Point.Y)
			partPoints = append(partPoints, point)
			if x > point.X {
				x = point.X
			}
			if y > point.Y {
				y = point.Y
			}
			if x2 < point.X {
				x2 = point.X
			}
			if y2 < point.Y {
				y2 = point.Y
			}
		}
	}
	if len(partPoints) < 5 {
		return ZR
	}
	// ------ Adjust heuristically +
	// if face points are detcted, adjust y value
	nosePart, foundNose := partsMap[CocoPartNose]
	if !foundNose {
		return ZR
	}
	var size float64
	neckPart, foundNeck := partsMap[CocoPartNeck]
	if foundNeck {
		size = math.Max(size, imgH*(neckPart.Point.Y-nosePart.Point.Y)*0.8)
	}
	rEyePart, foundREye := partsMap[CocoPartREye]
	lEyePart, foundLEye := partsMap[CocoPartLEye]
	if foundREye && foundLEye {
		size = math.Max(size, imgW*(rEyePart.Point.X-lEyePart.Point.X)*2.0)
		eyeX := math.Pow(rEyePart.Point.X-lEyePart.Point.X, 2.0)
		eyeY := math.Pow(rEyePart.Point.Y-lEyePart.Point.Y, 2.0)
		size = math.Max(size, imgW*math.Sqrt(eyeX+eyeY)*2.0)
	}
	if mode == 1 && !foundREye && !foundLEye {
		return ZR
	}
	rEarPart, foundREar := partsMap[CocoPartREar]
	lEarPart, foundLEar := partsMap[CocoPartLEar]
	if foundREar && foundLEar {
		size = math.Max(size, imgW*(rEarPart.Point.X-lEarPart.Point.X)*1.6)
	}
	if size <= 1e-15 {
		return ZR
	}
	if !foundREye && foundLEye {
		x = nosePart.Point.X*imgW - math.Floor(size/3)*2
	} else if foundREye && !foundLEye {
		x = nosePart.Point.X*imgW - math.Floor(size/3)
	} else { // foundREye && foundLEye
		x = nosePart.Point.X*imgW - math.Floor(size/2)
	}
	x2 = x + size
	if mode == 0 {
		y = nosePart.Point.Y*imgH - math.Floor(size/3)
	} else {
		y = nosePart.Point.Y*imgH - math.Round(size/2*1.2)
	}
	y2 = y + size
	// fit into the image frame
	x = math.Max(0, x)
	y = math.Max(0, y)
	x2 = math.Min(imgW-x, x2-x) + x
	y2 = math.Min(imgH-y, y2-y) + y
	if RoundInt(x2-x) == 0 || RoundInt(y2-y) == 0 {
		return ZR
	}
	if mode == 0 {
		return Rect(
			RoundInt((x+x2)/2),
			RoundInt((y+y2)/2),
			RoundInt(x2-x),
			RoundInt(y2-y),
		)
	}
	return Rect(
		RoundInt(x),
		RoundInt(y),
		RoundInt(x2-x),
		RoundInt(y2-y),
	)
}

// GetUpperBodyBox returns upper body box compared to img size (w, h)
func (h Human) GetUpperBodyBox(imgW float64, imgH float64) Rectangle {
	parts := make([]BodyPart, 0, TotalBodyParts)
	partsMap := make(map[CocoPart]BodyPart, TotalBodyParts)
	partPoints := make([]Point, 0, len(CoordPartsMap))
	var (
		x  float64
		y  float64
		x2 float64
		y2 float64
	)
	for _, part := range h.Parts {
		if part.Score <= ThresholdPartConfidence {
			continue
		}
		parts = append(parts, part)
		partsMap[part.Part] = part
		if _, found := CoordPartsMap[part.Part]; found {
			point := Pt(imgW*part.Point.X, imgH*part.Point.Y)
			partPoints = append(partPoints, point)
			if x > point.X {
				x = point.X
			}
			if y > point.Y {
				y = point.Y
			}
			if x2 < point.X {
				x2 = point.X
			}
			if y2 < point.Y {
				y2 = point.Y
			}
		}
	}
	if len(partPoints) < 5 {
		return ZR
	}
	// ------ Adjust heuristically +
	// if face points are detcted, adjust y value
	_, foundNose := partsMap[CocoPartNose]
	// var torsoHeight float64
	neckPart, foundNeck := partsMap[CocoPartNeck]
	if foundNose && foundNeck {
		y -= (neckPart.Point.Y*imgH - y) * 0.8
		// torsoHeight = math.Max(0, imgH*(neckPart.Point.Y-nosePart.Point.Y)*2.5)
	}
	// by using shoulder position, adjust width
	lShoulderPart, foundLShoulder := partsMap[CocoPartLShoulder]
	rShoulderPart, foundRShoulder := partsMap[CocoPartRShoulder]
	if foundLShoulder && foundRShoulder {
		halfW := x2 - x
		dx := halfW * 0.15
		x -= dx
		x2 += dx
	} else if foundNeck {
		if foundLShoulder && !foundRShoulder {
			halfW := math.Abs(lShoulderPart.Point.X-neckPart.Point.X) * imgW * 1.15
			x = math.Min(neckPart.Point.X*imgW-halfW, x)
			x2 = math.Max(neckPart.Point.X*imgW+halfW, x2)
		} else if !foundLShoulder && foundRShoulder {
			halfW := math.Abs(rShoulderPart.Point.X-neckPart.Point.X) * imgW * 1.15
			x = math.Min(neckPart.Point.X*imgW-halfW, x)
			x2 = math.Max(neckPart.Point.X*imgW+halfW, x2)
		}
	}

	// ------ Adjust heuristically -
	// fit into the image frame
	x = math.Max(0, x)
	y = math.Max(0, y)
	x2 = math.Min(imgW-x, x2-x) + x
	y2 = math.Min(imgH-y, y2-y) + y

	if RoundInt(x2-x) == 0 || RoundInt(y2-y) == 0 {
		return ZR
	}
	return Rect(
		RoundInt((x+x2)/2),
		RoundInt((y+y2)/2),
		RoundInt(x2-x),
		RoundInt(y2-y),
	)
}
