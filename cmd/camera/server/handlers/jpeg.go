package handlers

import (
	"log"
	"net/http"

	"github.com/bububa/camera"
	"github.com/bububa/camera/image"
	"github.com/bububa/openpose"
)

// JPEG handler.
type JPEG struct {
	e   *openpose.PoseEstimator
	cam *camera.Camera
}

// NewJPEG returns new JPEG handler.
func NewJPEG(e *openpose.PoseEstimator, cam *camera.Camera) *JPEG {
	return &JPEG{e, cam}
}

// ServeHTTP handles requests on incoming connections.
func (s *JPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Connection", "close")
	w.Header().Add("Cache-Control", "no-store, no-cache")
	w.Header().Add("Content-Type", "image/jpeg")

	img, err := s.cam.Read()
	if err != nil {
		log.Printf("jpeg: read: %v", err)
		return
	}
	if s.e != nil {
		if humans, err := s.e.Estimate(img, openpose.ModelSizeFaster); err == nil {
			img = openpose.DrawHumans(img, humans, 3)
		}
	}

	if err := image.NewEncoder(w).Encode(img); err != nil {
		log.Printf("jpeg: encode: %v", err)
		return
	}
}
