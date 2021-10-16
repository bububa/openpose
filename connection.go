package openpose

import (
	"fmt"
	"image"
	"sync"

	comb "github.com/bububa/openpose/combinations"
)

// Connection represents coco part connection
type Connection struct {
	Score       float32
	Coords      [2]image.Point
	Idx         [2]int
	Parts       [2]CocoPart
	Scores      [2]float32
	NormPadding Size
	UPartIdx    map[string]struct{}
}

// GetUPartIdx generate uniq partidx string for merge connections
func (c Connection) GetUPartIdx() map[string]struct{} {
	ret := make(map[string]struct{}, 2)
	key1 := fmt.Sprintf("%d-%d-%d", c.Coords[0].X, c.Coords[0].Y, c.Parts[0])
	key2 := fmt.Sprintf("%d-%d-%d", c.Coords[1].X, c.Coords[1].Y, c.Parts[1])
	ret[key1] = struct{}{}
	ret[key2] = struct{}{}
	return ret
}

// ToBodyPart returns BodyParts based on Connection
func (c Connection) ToBodyParts(parts *[2]BodyPart, rows float64, cols float64) {
	parts[0] = BodyPart{
		Part: c.Parts[0],
		Point: Pt(
			float64(c.Coords[0].X)/float64(cols)/c.NormPadding.W,
			float64(c.Coords[0].Y)/float64(rows)/c.NormPadding.H,
		),
		Score: c.Scores[0],
	}
	parts[1] = BodyPart{
		Part: c.Parts[1],
		Point: Pt(
			float64(c.Coords[1].X)/float64(cols)/c.NormPadding.W,
			float64(c.Coords[1].Y)/float64(rows)/c.NormPadding.H,
		),
		Score: c.Scores[1],
	}
}

func connectionsCouldMerge(c1 Connection, c2 Connection) bool {
	for k, _ := range c2.UPartIdx {
		if _, found := c1.UPartIdx[k]; found {
			return true
		}
	}
	return false
}

func connectionsIterProduct(mp map[string][]Connection, keys []string) <-chan [2]Connection {
	ch := make(chan [2]Connection)
	pool := &sync.Pool{
		New: func() interface{} {
			return [2]Connection{}
		},
	}
	go func(pool *sync.Pool) {
		for _, c1 := range mp[keys[0]] {
			for _, c2 := range mp[keys[1]] {
				val := pool.Get().([2]Connection)
				val[0] = c1
				val[1] = c2
				pool.Put(val)
				ch <- val
			}
		}
		close(ch)
	}(pool)
	return ch
}

func joinConnections(connections []Connection) map[string][]Connection {
	connectedByHuman := make(map[string][]Connection, len(connections))
	combineKeys := make([]interface{}, 0, len(connectedByHuman))
	for idx, conn := range connections {
		key := fmt.Sprintf("human-%d", idx)
		connectedByHuman[key] = []Connection{conn}
		combineKeys = append(combineKeys, key)
	}
	combinationProducts, err := comb.NewCombination(combineKeys, 2)
	if err != nil {
		return connectedByHuman
	}
	noMergeCache := make(map[string]map[string]struct{}, len(connections))
	keyPool := &sync.Pool{
		New: func() interface{} {
			return make([]string, 2)
		},
	}
	for {
		var isMerged bool
		for combinationProducts.Next() {
			values := combinationProducts.Value()
			if len(values) != 2 {
				break
			}
			keys := keyPool.Get().([]string)
			keys[0] = values[0].(string)
			keys[1] = values[1].(string)
			keyPool.Put(keys)
			if keys[0] == keys[1] {
				continue
			}
			if _, found := connectedByHuman[keys[0]]; !found {
				continue
			}
			if _, found := connectedByHuman[keys[1]]; !found {
				continue
			}
			if mp, found := noMergeCache[keys[0]]; found {
				if _, found := mp[keys[1]]; found {
					continue
				}
			} else {
				noMergeCache[keys[0]] = make(map[string]struct{}, len(keys))
			}
			for cs := range connectionsIterProduct(connectedByHuman, keys) {
				if connectionsCouldMerge(cs[0], cs[1]) {
					isMerged = true
					connectedByHuman[keys[0]] = append(connectedByHuman[keys[0]], connectedByHuman[keys[1]]...)
					delete(connectedByHuman, keys[1])
					break
				}
			}
			if isMerged {
				delete(noMergeCache, keys[0])
				break
			}
		}
		if !isMerged {
			break
		}
	}
	return connectedByHuman
}

func connectionsToHuman(human *Human, connections []Connection, heatMatRows float64, heatMatCols float64) {
	partsPool := &sync.Pool{
		New: func() interface{} {
			return [2]BodyPart{}
		},
	}
	for _, c := range connections {
		parts := partsPool.Get().([2]BodyPart)
		c.ToBodyParts(&parts, heatMatRows, heatMatCols)
		partsPool.Put(parts)
		human.Parts[parts[0].Part] = parts[0]
		human.Parts[parts[1].Part] = parts[1]
		human.Score += c.Score + parts[1].Score
	}
	human.Score = human.Score / float32(human.PartCount())
}

func connectionsToHumans(connections []Connection, heatMatRows float64, heatMatCols float64) []Human {
	connectedByHuman := joinConnections(connections)
	humans := make([]Human, 0, len(connectedByHuman))
	humanPool := &sync.Pool{
		New: func() interface{} {
			return NewHuman()
		},
	}
	for _, conns := range connectedByHuman {
		// reject by subset count
		if len(conns) < MinSubsetCnt {
			continue
		}
		// reject by subset max score
		var maxScore float32
		for _, conn := range conns {
			if maxScore < conn.Score {
				maxScore = conn.Score
			}
		}
		if maxScore < MinSubsetScore {
			continue
		}
		h := humanPool.Get().(*Human)
		h.Reset()
		connectionsToHuman(h, conns, heatMatRows, heatMatCols)
		humanPool.Put(h)
		if h.Score < ThresholdHumanScore {
			continue
		}
		humans = append(humans, *h)
	}
	return humans
}
