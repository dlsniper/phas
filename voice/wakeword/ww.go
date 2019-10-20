//    Copyright 2019 Florin Pățan
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
	"log"
	"strings"

	"github.com/charithe/porcupine-go"
	"github.com/gordonklaus/portaudio"
)

//Listener handles the wakeword detection
type Listener struct {
	p porcupine.Porcupine
}

//Listen will process the user voice until a specific wakeword is detected
func (l *Listener) Listen(ctx context.Context) (string, context.Context) {
	frameSize := porcupine.FrameLength()
	audioFrame := make([]int16, frameSize)
	stream, err := portaudio.OpenDefaultStream(1, 0, 16000, frameSize, audioFrame)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	if err := stream.Start(); err != nil {
		log.Fatalln(err)
	}

	log.Println("listening for the special keywords...")

	for {
		if err := stream.Read(); err != nil {
			log.Fatalln(err)
		}

		word, err := l.p.Process(audioFrame)
		if err != nil {
			log.Fatalf("error: %+v\n", err)
		}

		if word != "" {
			// TODO replace the context key with something better, not a string
			return strings.TrimSpace(strings.ToLower(word)), context.WithValue(ctx, "phasReceivedWakeWord", word)
		}
	}
}

//NewListener creates a new service to listen for the registered wakewords
func NewListener(modelPath string, keywords []*porcupine.Keyword) *Listener {
	p, err := porcupine.New(modelPath, keywords...)
	if err != nil {
		log.Fatalf("failed to initialize porcupine: %+v\n", err)
	}
	return &Listener{
		p: p,
	}
}
