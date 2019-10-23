//    Copyright 2021 Florin Pățan
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package sentry

import (
	"context"
	"image"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dlsniper/phas/tts"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

//Service watches the house for stuff
type Service struct {
	camera      int
	sensibility float64
	started     bool
	once        sync.Once
	gwait, wait chan struct{}
	state       chan struct{}
}

var streamingServerAddr = ":42080"

//New create a new Service
func New(camera int, sensibility float64, gwait chan struct{}, callback func(sinceLastAlarm float64)) *Service {
	res := &Service{
		camera:      camera,
		sensibility: sensibility,
		gwait:       gwait,
		wait:        make(chan struct{}),
		state:       make(chan struct{}),
	}

	res.once.Do(func() {
		go func() {
			var lastAlarm time.Time
			for {
				<-res.state
				now := time.Now()
				callback(now.Sub(lastAlarm).Seconds())
				lastAlarm = now
			}
		}()
	})

	return res
}

//Toggle the Service state to on or off
func (s *Service) Toggle(ctx context.Context, ttsService *tts.Service, desiredSentryMode bool) {
	if desiredSentryMode && !s.started {
		go func() {
			ttsService.Speak(ctx, "Sentry mode activated!")
			s.Start(s.camera, s.sensibility)
		}()
		s.started = true
	} else if !desiredSentryMode && s.started {
		s.wait <- struct{}{}
		ttsService.Speak(ctx, "Sentry mode turned off!")
		s.started = false
	}
}

//Start the Service state
func (s *Service) Start(cameraID int, detectionSensibility float64) {
	webcam, err := gocv.OpenVideoCapture(cameraID)
	if err != nil {
		log.Printf("Error opening video capture device: %v\n", cameraID)
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	// create the mjpeg stream
	stream := mjpeg.NewStream()

	go func() {
		// start http server
		http.Handle("/sentry/camera", stream)
		server := &http.Server{Addr: streamingServerAddr}

		go func() {
			<-s.wait
			_ = server.Close()
		}()

		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Failed to start streaming server with error %v\n", err)
			return
		}
	}()

	for {
		if ok := webcam.Read(&img); !ok {
			log.Printf("Problem reading the camera: %v\n", cameraID)
		}
		if img.Empty() {
			continue
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf)

		mog2.Apply(img, &imgDelta)
		gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

		// To ensure we call the k.Close method, we run this in a closure
		func() {
			k := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
			defer func(kernel *gocv.Mat) {
				_ = kernel.Close()
			}(&k)
			gocv.Dilate(imgThresh, &imgThresh, k)
		}()

		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

		for i := 0; i < contours.Size(); i++ {
			ca := gocv.ContourArea(contours.At(i))
			if ca > detectionSensibility {
				s.state <- struct{}{}
				break
			}
		}

		select {
		case <-s.gwait:
			return
		case <-s.wait:
			return
		default:
			continue
		}
	}
}
