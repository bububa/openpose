package openpose

import (
	"math"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

// nonMaxSuppression uppress performs non-max suppression on a sorted list of detections.
func nonMaxSuppression(imap [][]float32, scale float64, threshold float32) ([][]float32, error) {
	l := len(imap)
	bbox, scores := generateBbox(imap, scale, threshold)
	tbbox, err := tf.NewTensor(bbox)
	if err != nil {
		return nil, err
	}
	tscore, err := tf.NewTensor(scores)
	if err != nil {
		return nil, err
	}

	s := op.NewScope()
	pbbox := op.Placeholder(s.SubScope("bbox"), tf.Float, op.PlaceholderShape(tf.MakeShape(-1, 4)))
	pscore := op.Placeholder(s.SubScope("score"), tf.Float, op.PlaceholderShape(tf.MakeShape(-1)))

	out := op.NonMaxSuppression(s, pbbox, pscore, op.Const(s.SubScope("max_len"), int32(l)), op.NonMaxSuppressionIouThreshold(threshold))

	outs, err := runScope(s, map[tf.Output]*tf.Tensor{pbbox: tbbox, pscore: tscore}, []tf.Output{out})
	if err != nil {
		return nil, err
	}
	ret := make([][]float32, l)
	for i, v := range imap {
		ret[i] = make([]float32, len(v))
	}
	pick := outs[0]
	if pick != nil {
		if idx, ok := pick.Value().([]int32); ok {
			rowLen := len(imap[0])
			for _, i := range idx {
				x := int(i)
				y := x / rowLen
				if x >= rowLen {
					x %= rowLen
				}
				// log.Printf("i:%d, x:%d, y:%d, score:%f, imap:%f\n", i, x, y, scores[i], imap[y][x])
				ret[y][x] = scores[i]
			}
		}
	}
	return ret, nil
}

func resizeArea(imgs []*tf.Tensor, upsampleSize []int32) ([]*tf.Tensor, error) {
	ret := make([]*tf.Tensor, 0, len(imgs))
	for _, img := range imgs {
		s := op.NewScope()
		input := op.Placeholder(s.SubScope("image"), tf.Float, op.PlaceholderShape(tf.MakeShape(img.Shape()...)))
		upsample, err := tf.NewTensor(upsampleSize)
		if err != nil {
			return nil, err
		}
		size := op.Placeholder(s.SubScope("upsample"), tf.Int32, op.PlaceholderShape(tf.MakeShape(2)))
		out := op.ResizeArea(s, input, size)
		outs, err := runScope(s, map[tf.Output]*tf.Tensor{input: img, size: upsample}, []tf.Output{out})
		if err != nil {
			return nil, err
		}
		ret = append(ret, outs[0])
	}
	return ret, nil
}

func runScope(s *op.Scope, inputs map[tf.Output]*tf.Tensor, outputs []tf.Output) ([]*tf.Tensor, error) {
	graph, err := s.Finalize()
	if err != nil {
		return nil, err
	}

	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.Run(inputs, outputs, nil)
}

func generateBbox(imap [][]float32, scale float64, threshold float32) ([][]float32, []float32) {
	const (
		Stride   = 2.0
		CellSize = 12.0
	)
	var (
		l      = len(imap)
		bbox   = make([][]float32, 0, l*len(imap[0]))
		scores = make([]float32, 0, l*len(imap[0]))
	)

	for i, x := range imap {
		for j, y := range x {
			n := []float32{float32(math.Floor((Stride*float64(j)+1.0)/scale + 0.5)),
				float32(math.Floor((Stride*float64(i)+1.0)/scale + 0.5)),
				float32(math.Floor((Stride*float64(j)+1.0+CellSize)/scale + 0.5)),
				float32(math.Floor((Stride*float64(i)+1.0+CellSize)/scale + 0.5)),
			}
			bbox = append(bbox, n)
			scores = append(scores, y)
			/*
				if y > threshold {
					scores = append(scores, y)
				} else {
					scores = append(scores, 0)
				}
			*/
		}
	}

	return bbox, scores
}
