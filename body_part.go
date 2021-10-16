package openpose

import "fmt"

// BodyPart represents body part
type BodyPart struct {
	// Part CocoPart
	Part CocoPart
	// Point coordinate of body part
	Point Point
	// Score confidence score
	Score float32
}

// NewBodyPart returns a new BodyPart
func NewBodyPart(part CocoPart, point Point, score float32) BodyPart {
	return BodyPart{
		Part:  part,
		Point: point,
		Score: score,
	}
}

// String return BodyPart string representation
func (b BodyPart) String() string {
	return fmt.Sprintf("BodyPart:%d-(%.2f, %.2f) score:%.4f", b.Part, b.Point.X, b.Point.Y, b.Score)
}

// BodyPartID returns a BodyPart.ID with given CocoPart and index
func BodyPartID(humanID int, part CocoPart) string {
	return fmt.Sprintf("%d-%d", humanID, part)
}

// BodyPartPair ...
type BodyPartPair struct {
	Parts  [2]CocoPart
	Points [2]Point
	Score  float64
}

// BodyPartsPoints return body part points
func BodyPartsPoints(parts []BodyPart) ([]Point, []bool) {
	mp := make(map[CocoPart]BodyPart, len(parts))
	for _, part := range parts {
		mp[part.Part] = part
	}
	points := make([]Point, 0, len(PartPairs))
	visibility := make([]bool, 0, len(PartPairs))
	for _, pair := range PartPairs {
		if part, found := mp[pair.CocoPart]; found {
			points = append(points, part.Point)
			visibility = append(visibility, true)
			continue
		}
		points = append(points, ZP)
		visibility = append(visibility, false)
	}
	return points, visibility
}
