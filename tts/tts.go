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

package tts

import (
	"bytes"
	"context"
	"io"
	"log"

	"cloud.google.com/go/texttospeech/apiv1"
	"github.com/hajimehoshi/oto"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type config struct {
	audioConfig *texttospeechpb.AudioConfig
	voice       *texttospeechpb.VoiceSelectionParams
}

//Service processes the text to speech content transformation
type Service struct {
	service *texttospeech.Client
	config  *config
	player  *oto.Player
}

//Speak will read the text back to the user
func (s Service) Speak(ctx context.Context, text string) {
	req := s.newRequest(text)

	resp, err := s.service.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.AudioContent == nil {
		log.Fatalln("nil audio response from GCP...")
	}

	response := bytes.NewReader(resp.AudioContent)
	readBytes := int64(0)
	for {
		//goland:noinspection GoShadowedVar
		n, err := io.Copy(s.player, response)
		readBytes += n
		// It seems that sometimes the io.EOF is not enough to stop
		// so we need to keep track of the read bytes...
		if err == io.EOF || int(readBytes) >= len(resp.AudioContent) {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (s *Service) newRequest(text string) texttospeechpb.SynthesizeSpeechRequest {
	return texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		AudioConfig: s.config.audioConfig,
		Voice:       s.config.voice,
	}
}

//New creates a new text to speech service
func New(service *texttospeech.Client) *Service {
	playerCtx, err := oto.NewContext(24000, 1, 2, 8192)
	if err != nil {
		log.Fatalln(err)
	}

	player := playerCtx.NewPlayer()

	return &Service{
		service: service,
		config: &config{
			audioConfig: &texttospeechpb.AudioConfig{
				AudioEncoding: texttospeechpb.AudioEncoding_LINEAR16,
			},
			voice: &texttospeechpb.VoiceSelectionParams{
				Name:         "en-US-Wavenet-D",
				LanguageCode: "en-US",
			},
		},
		player: player,
	}
}
