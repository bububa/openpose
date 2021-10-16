package openpose

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"math"

	"github.com/disintegration/imaging"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

// ImagePreprocess preprocess image for model
func ImagePreprocess(img image.Image, modelSize ModelSize, sharpenSigma float64) (image.Image, Size) {
	imgW := float64(img.Bounds().Max.X)
	imgH := float64(img.Bounds().Max.Y)
	targetW := float64(modelSize[0])
	targetH := float64(modelSize[1])
	out := resizeImage(img, imgW, imgH, targetW, targetH)
	out, normPadding := padImage(out, targetW, targetH)
	if sharpenSigma > 1e-15 {
		out = imaging.Sharpen(out, sharpenSigma)
	}
	return out, normPadding
}

func resizeImage(img image.Image, sourceW float64, sourceH float64, targetW float64, targetH float64) image.Image {
	// Compute correct scale for resizing operation
	var scale float64
	targetRatio := targetH / targetW
	sourceRatio := sourceH / sourceW
	if targetRatio > sourceRatio {
		scale = targetW / sourceW
	} else {
		scale = targetH / sourceH
	}
	return imaging.Resize(img, int(sourceW*scale), int(sourceH*scale), imaging.Linear)
}

func padImage(img image.Image, targetW float64, targetH float64) (image.Image, Size) {
	imgW := float64(img.Bounds().Max.X)
	imgH := float64(img.Bounds().Max.Y)
	normPadding := Size{
		W: imgW / targetW,
		H: imgH / targetH,
	}
	out := image.NewRGBA(image.Rect(0, 0, int(targetW), int(targetH)))
	gc := draw2dimg.NewGraphicContext(out)
	gc.DrawImage(img)
	return out, normPadding
}

func makeTensorFromImage(img image.Image) (*tf.Tensor, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}
	tensor, err := tf.NewTensor(buf.String())
	if err != nil {
		return nil, err
	}
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	graph, input, output, err := makeTransformImageGraph(int32(w), int32(h))
	if err != nil {
		return nil, err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	out, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	if len(out) < 1 || len(out[0].Value().([][][][]float32)) < 1 {
		return nil, errors.New("invalid output")
	}
	mean, std := meanStd(out[0].Value().([][][][]float32)[0])
	return preWhitenImage(out[0], mean, std)
}

// Creates a graph to decode, rezise and normalize an image
func makeTransformImageGraph(width int32, height int32) (graph *tf.Graph, input, output tf.Output, err error) {
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	// Decode PNG or JPEG
	decode := op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	// Div and Sub perform (value-Mean)/Scale for each pixel
	output = op.ResizeBilinear(s,
		// Create a batch containing a single image
		op.ExpandDims(s,
			// Use decoded pixel values
			op.Cast(s, decode, tf.Float),
			op.Const(s.SubScope("make_batch"), int32(0))),
		op.Const(s.SubScope("size"), []int32{height, width}),
	)
	graph, err = s.Finalize()
	return graph, input, output, err
}

func preWhitenImage(img *tf.Tensor, mean, std float32) (*tf.Tensor, error) {
	s := op.NewScope()
	pimg := op.Placeholder(s, tf.Float, op.PlaceholderShape(tf.MakeShape(1, -1, -1, 3)))

	out := op.Mul(s, op.Sub(s, pimg, op.Const(s.SubScope("mean"), mean)),
		op.Const(s.SubScope("scale"), float32(1.0)/std))
	outs, err := runScope(s, map[tf.Output]*tf.Tensor{pimg: img}, []tf.Output{out})
	if err != nil {
		return nil, err
	}

	return outs[0], nil
}

func convertValue(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}

func meanStd(img [][][]float32) (mean float32, std float32) {
	count := len(img) * len(img[0]) * len(img[0][0])
	for _, x := range img {
		for _, y := range x {
			for _, z := range y {
				mean += z
			}
		}
	}
	mean /= float32(count)

	for _, x := range img {
		for _, y := range x {
			for _, z := range y {
				std += (z - mean) * (z - mean)
			}
		}
	}

	xstd := math.Sqrt(float64(std) / float64(count-1))
	minstd := 1.0 / math.Sqrt(float64(count))
	if xstd < minstd {
		xstd = minstd
	}

	std = float32(xstd)
	return
}

func DrawHumans(img image.Image, humans []Human, strokeWidth float64) image.Image {
	imgW := float64(img.Bounds().Max.X)
	imgH := float64(img.Bounds().Max.Y)
	//log.Printf("w:%f, h:%f\n", imgW, imgH)
	out := image.NewRGBA(img.Bounds())
	gc := draw2dimg.NewGraphicContext(out)
	gc.DrawImage(img)
	startPart := CocoPartNose
	maxPart := CocoPartLEar
	for _, human := range humans {
		// draw points
		part := startPart
		centers := make(map[CocoPart]Point, maxPart)
		for part <= maxPart {
			if !human.HasPart(part) {
				part++
				continue
			}
			coord := human.Parts[part].Point
			center := Pt(coord.X*imgW+0.5, coord.Y*imgH+0.5)
			//log.Printf("%d, %d, (%d-%d)\n", humanID, part, int(center.X), int(center.Y))
			centers[part] = center
			cocoColor := CocoColors[part]
			partColor := color.RGBA{cocoColor[0], cocoColor[1], cocoColor[2], 255}
			gc.SetFillColor(partColor)
			gc.SetStrokeColor(partColor)
			gc.SetLineWidth(strokeWidth * 0.5)
			draw2dkit.Circle(gc, center.X, center.Y, strokeWidth)
			gc.FillStroke()
			part++
		}
		// draw lines
		for idx, pair := range CocoPairsRender {
			if !human.HasPart(pair[0]) || !human.HasPart(pair[1]) {
				continue
			}
			cocoColor := CocoColors[idx]
			lineColor := color.RGBA{cocoColor[0], cocoColor[1], cocoColor[2], 255}
			//log.Printf("%d, %d-%d, (%d, %d), (%d, %d)\n", humanID, pair[0], pair[1], int(centers[pair[0]].X), int(centers[pair[0]].Y), int(centers[pair[1]].X), int(centers[pair[1]].Y))
			gc.SetStrokeColor(lineColor)
			gc.SetFillColor(lineColor)
			gc.SetLineWidth(2)
			gc.BeginPath()
			gc.MoveTo(centers[pair[0]].X, centers[pair[0]].Y)
			gc.LineTo(centers[pair[1]].X, centers[pair[1]].Y)
			gc.Close()
			gc.FillStroke()
		}
	}
	return out
}
