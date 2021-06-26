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

package wakeword

import (
	"context"
	"encoding/binary"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Picovoice/porcupine/binding/go"
	"github.com/gen2brain/malgo"
)

//Keyword is the wakeword keyword
type Keyword struct {
	Name        string
	FilePath    string
	Sensitivity float32
}

//Listener handles the wakeword detection
type Listener struct {
	p   porcupine.Porcupine
	kws []string
}

func (l *Listener) onAudioData(onData chan string) func([]byte, []byte, uint32) {
	var shortBufIndex, shortBufOffset int
	shortBuf := make([]int16, porcupine.FrameLength)
	return func(_, pSample []byte, frameCount uint32) {
		for i := 0; i < len(pSample); i += 2 {
			shortBuf[shortBufIndex+shortBufOffset] = int16(binary.LittleEndian.Uint16(pSample[i : i+2]))
			shortBufOffset++

			if shortBufIndex+shortBufOffset == porcupine.FrameLength {
				shortBufIndex = 0
				shortBufOffset = 0
				//goland:noinspection GoShadowedVar
				kidx, _ := l.p.Process(shortBuf)
				if kidx >= 0 && kidx < len(l.kws) {
					onData <- l.kws[kidx]
					return
				}
			}
		}

		shortBufIndex += shortBufOffset
		shortBufOffset = 0
	}
}

//Listen will process the user voice until a specific wakeword is detected
func (l *Listener) Listen(ctx context.Context) (string, context.Context) {
	var backends []malgo.Backend = nil
	var sampleRate = uint32(porcupine.SampleRate)
	if runtime.GOOS == "windows" {
		backends = []malgo.Backend{malgo.BackendDsound}
	} else if runtime.GOOS == "linux" {
		backends = []malgo.Backend{malgo.BackendAlsa}
		sampleRate = uint32(porcupine.SampleRate / 2)
	}

	mctx, err := malgo.InitContext(backends, malgo.ContextConfig{}, func(message string) {
		log.Println(message)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = mctx.Uninit()
		mctx.Free()
	}()

	deviceConfig := func() malgo.DeviceConfig {
		deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.SampleRate = sampleRate
		deviceConfig.Alsa.NoMMap = 1
		return deviceConfig
	}()

	onData := make(chan string)
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: l.onAudioData(onData),
	}
	device, err := malgo.InitDevice(mctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		log.Fatalln(err)
	}
	defer device.Uninit()

	err = l.p.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = l.p.Delete()
	}()

	err = device.Start()
	if err != nil {
		log.Fatalln(err)
	}
	defer func(device *malgo.Device) {
		_ = device.Stop()
	}(device)

	select {
	case <-ctx.Done():
		return "", ctx
	case kw := <-onData:
		return kw, ctx
	}
}

//NewListener creates a new service to listen for the registered wakewords
func NewListener(modelPath string, keywords []Keyword) *Listener {
	modelPath, err := filepath.Abs(modelPath)
	if err != nil {
		log.Fatalln(err)
	}
	if //goland:noinspection GoShadowedVar
	_, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Fatalf("Could not find model file at %s", modelPath)
	}

	p := porcupine.Porcupine{}
	p.ModelPath = modelPath
	var kws []string
	for _, kw := range keywords {
		keywordPath, _ := filepath.Abs(kw.FilePath)
		if //goland:noinspection GoShadowedVar
		_, err := os.Stat(keywordPath); os.IsNotExist(err) {
			log.Fatalf("Could not find keyword file at %s", keywordPath)
		}
		kws = append(kws, kw.Name)
		p.KeywordPaths = append(p.KeywordPaths, keywordPath)
		p.Sensitivities = append(p.Sensitivities, kw.Sensitivity)
	}

	return &Listener{
		kws: kws,
		p:   p,
	}
}
