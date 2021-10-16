// go:build opencv

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/bububa/camera"
	"github.com/bububa/openpose"
	"github.com/bububa/openpose/cmd/camera/server"
)

var (
	opts      camera.Options
	estimator *openpose.PoseEstimator
	bind      string
	modelPath string
	modelType string
)

func init() {
	flag.IntVar(&opts.Index, "index", 0, "Camera index")
	flag.IntVar(&opts.Delay, "delay", 10, "Delay between frames, in milliseconds")
	flag.Float64Var(&opts.Width, "width", 640, "Frame width")
	flag.Float64Var(&opts.Height, "height", 480, "Frame height")
	flag.StringVar(&bind, "bind", ":8080", "set server bind")
	flag.StringVar(&modelPath, "model", "", "set openpose model path")
	flag.StringVar(&modelType, "model-type", "mobilenet", "set openpose model type")
}

func setup() error {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	modelPath = cleanPath(wd, modelPath)
	mt := openpose.MobileNet
	if modelType == "cmu" {
		mt = openpose.CMU
	}
	estimator = openpose.NewPoseEstimator(
		modelPath,
		mt,
	)
	if err != nil {
		return err
	}
	return nil

}

func main() {
	if err := setup(); err != nil {
		log.Fatalln(err)
	}
	log.Println("getting device...")
	device, err := getDevice(opts)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("starting server...")
	cam := camera.NewCamera(device)
	srv := server.New(bind, estimator, cam)
	srv.SetFrameSize(opts.Width, opts.Height)
	srv.SetDelay(opts.Delay)

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()
	log.Printf("server started at %s\n", bind)
	<-exitCh
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); nil != err {
		log.Fatalf("server shutdown failed, err: %v\n", err)
	}
	log.Println("server gracefully shutdown")
}
func cleanPath(wd string, path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		return dir
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(dir, path[2:])
	}
	return filepath.Join(wd, path)
}
