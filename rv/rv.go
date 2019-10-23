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

package rv

import (
	"encoding/binary"
	"io"
	"log"
	"runtime"
	"time"

	"github.com/Picovoice/porcupine/binding/go"
	"github.com/gen2brain/malgo"
	"github.com/go-audio/wav"
	"github.com/orcaman/writerseeker"
)

//Service that handles the voice recording
type Service struct{}

func (s *Service) recordVoice() []byte {
	log.Println("recording voice")

	var backends []malgo.Backend = nil
	sampleRate := uint32(porcupine.SampleRate)
	if runtime.GOOS == "windows" {
		backends = []malgo.Backend{malgo.BackendDsound}
	} else if runtime.GOOS == "linux" {
		sampleRate = uint32(porcupine.SampleRate / 2)
		backends = []malgo.Backend{malgo.BackendAlsa}
	}

	ctx, err := malgo.InitContext(backends, malgo.ContextConfig{}, func(m string) {
		log.Println(m)
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := func() malgo.DeviceConfig {
		deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.SampleRate = sampleRate
		deviceConfig.Alsa.NoMMap = 1
		return deviceConfig
	}()

	ws := &writerseeker.WriterSeeker{}
	outputWav := wav.NewEncoder(ws, 16000, 16, 1, 1)

	var shortBufIndex, shortBufOffset int
	shortBuf := make([]int16, porcupine.FrameLength)
	onRecvFrames := func(_, in []byte, frameCount uint32) {
		for i := 0; i < len(in); i += 2 {
			shortBuf[shortBufIndex+shortBufOffset] = int16(binary.LittleEndian.Uint16(in[i : i+2]))
			shortBufOffset++

			if shortBufIndex+shortBufOffset == porcupine.FrameLength {
				shortBufIndex = 0
				shortBufOffset = 0
				for outputBufIndex := range shortBuf {
					er := outputWav.WriteFrame(shortBuf[outputBufIndex])
					if er != nil {
						log.Fatalln(err)
					}
				}
			}
		}
	}

	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if err != nil {
		log.Fatalln(err)
	}

	err = device.Start()
	if err != nil {
		log.Fatalln(err)
	}

	// Wait for the command to finish
	time.Sleep(4 * time.Second)

	device.Stop()
	outputWav.Close()

	out, _ := io.ReadAll(ws.Reader())
	return out
}

//Listen begins listening for user input
func (s *Service) Listen() []byte {
	return s.recordVoice()
}

//New creates a new voice recording service
func New() *Service {
	return &Service{}
}
