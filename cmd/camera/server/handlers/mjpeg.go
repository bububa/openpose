package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/bububa/camera"
	"github.com/bububa/camera/image"
	"github.com/bububa/openpose"
)

// MJPEG handler.
type MJPEG struct {
	e     *openpose.PoseEstimator
	cam   *camera.Camera
	delay int
}

// NewMJPEG returns new MJPEG handler.
func NewMJPEG(e *openpose.PoseEstimator, cam *camera.Camera, delay int) *MJPEG {
	return &MJPEG{e, cam, delay}
}

// ServeHTTP handles requests on incoming connections.
func (s *MJPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	mimeWriter := multipart.NewWriter(w)
	mimeWriter.SetBoundary("--boundary")

	w.Header().Add("Connection", "close")
	w.Header().Add("Cache-Control", "no-store, no-cache")
	w.Header().Add("Content-Type", fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary()))

	cn := w.(http.CloseNotifier).CloseNotify()

loop:
	for {
		select {
		case <-cn:
			break loop

		default:
			partHeader := make(textproto.MIMEHeader)
			partHeader.Add("Content-Type", "image/jpeg")

			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Printf("mjpeg: createPart: %v", err)
				continue
			}

			img, err := s.cam.Read()
			if err != nil {
				log.Printf("jpeg: read: %v", err)
				return
			}
			if s.e != nil {
				if humans, err := s.e.Estimate(img, openpose.ModelSizeFaster); err == nil {
					log.Printf("found:%d\n", len(humans))
					img = openpose.DrawHumans(img, humans, 3)
				} else {
					log.Println("nothing found")
				}
			}

			err = image.NewEncoder(partWriter).Encode(img)
			if err != nil {
				log.Printf("mjpeg: encode: %v", err)
				continue
			}

			time.Sleep(time.Duration(s.delay) * time.Millisecond)
		}
	}

	mimeWriter.Close()
}
