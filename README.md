# Golang lib for pose detection using tensorflow openpose

[![Go Reference](https://pkg.go.dev/badge/github.com/bububa/openpose.svg)](https://pkg.go.dev/github.com/bububa/openpose)
[![Go](https://github.com/bububa/openpose/actions/workflows/go.yml/badge.svg)](https://github.com/bububa/openpose/actions/workflows/go.yml)
[![goreleaser](https://github.com/bububa/openpose/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/bububa/openpose/actions/workflows/goreleaser.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/bububa/openpose.svg)](https://github.com/bububa/openpose)
[![GoReportCard](https://goreportcard.com/badge/github.com/bububa/openpose)](https://goreportcard.com/report/github.com/bububa/openpose)
[![GitHub license](https://img.shields.io/github/license/bububa/openpose.svg)](https://github.com/bububa/openpose/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/bububa/openpose.svg)](https://GitHub.com/bububa/openpose/releases/)

## Prerequest

1. libtensorfow 1.x
   Follow the instruction [Install TensorFlow for C](https://www.tensorflow.org/install/lang_c#macos)
2. download tenorflow model graph

- CMU model [Google Drive](https://drive.google.com/file/d/1OGucxzDKNtbLWmRM59_xJaG4odXDMIXA/view?usp=sharing)
- mobilenet large model [Google Drive](https://drive.google.com/file/d/1qQ1mz58G22DKhoxwP2wmSRZQLpRdgEmy/view?usp=sharing)
- mobilenet small model [Google Drive](https://drive.google.com/file/d/1n9-j-2KpBaqZyLDHc_mmaU8hodroL1mf/view?usp=sharing)
- mobilenet thin model [Google Drive](https://drive.google.com/drive/folders/1vZXHNDvpbdaAUYHsxjfh7Zgzm5slvU5P?usp=sharing)

## Demo

![demo screen capture](https://github.com/bububa/openpose/blob/main/demo.jpg?raw=true)

## Install

go get -u github.com/bububa/openpose

## Camera & Server

### Requirements

- [libjpeg-turbo](https://www.libjpeg-turbo.org/) (use `-tags jpeg` to build without `CGo`)
- On Linux/RPi native Go [V4L](https://github.com/korandiz/v4l) implementation is used to capture images.

### Use Opencv4

```bash
make cvcamera
```

### On linux/Pi

```bash
# use native Go V4L implementation is used to capture images
make linux_camera
```

### Use image/jpeg instead of libjpeg-turbo

use jpeg build tag to build with native Go `image/jpeg` instead of `libjpeg-turbo`

```bash
go build -o=./bin/cvcamera -tags=cv4,jpeg ./cmd/camera
```

### Usage as Server

```
Usage of camera:
  -bind string
	Bind address (default ":56000")
  -delay int
	Delay between frames, in milliseconds (default 10)
  -width float
	Frame width (default 640)
  -height float
	Frame height (default 480)
  -index int
	Camera index
  -model string
    mode path
```

## User as lib

```golang
import (
    "github.com/bububa/openpose"
)

func main() {
	t := openpose.NewPoseEstimator(modelPath, openpose.MobileNet)
	wd, _ := os.Getwd()
	img, err := loadImage("./golf.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	modelSize := openpose.ModelSizeFaster
	sharpenSigma := 0.0
	t.SetSharpenSigma(sharpenSigma)
	humans, err := t.Estimate(img, modelSize)
	if err != nil {
		log.Fatalln(err)
	}
	outImg := openpose.DrawHumans(img, humans, 2)
    saveImage(outImg, "./out/jpg")
}
```
