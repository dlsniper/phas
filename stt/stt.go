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

package stt

import (
	"context"
	"log"

	"cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

//Service that handles the speech to text conversion
type Service struct {
	service *speech.Client
	config  *speechpb.RecognitionConfig
}

func (s *Service) newRequest(content []byte) *speechpb.RecognizeRequest {
	return &speechpb.RecognizeRequest{
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: content,
			},
		},
		Config: s.config,
	}
}

func (s *Service) command(resp *speechpb.RecognizeResponse) string {
	command := ""
	for idx := range resp.Results {
		command += resp.Results[idx].Alternatives[0].Transcript
	}
	return command
}

//Process processes the incoming voice audio content and transforms it to text
func (s *Service) Process(ctx context.Context, content []byte) string {
	req := s.newRequest(content)

	resp, err := s.service.Recognize(ctx, req)
	if err != nil {
		log.Fatalln(err)
	}

	return s.command(resp)
}

//New creates a new speech to text service
func New(speechService *speech.Client) *Service {
	return &Service{
		service: speechService,
		config: &speechpb.RecognitionConfig{
			LanguageCode: "en-US",
			Model:        "command_and_search",
			Encoding:     speechpb.RecognitionConfig_ENCODING_UNSPECIFIED,
		},
	}
}
