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

package commands

import (
	"context"
	"log"
	"strings"

	"github.com/dlsniper/phas/commands/intents"
	"github.com/dlsniper/phas/tts"
)

//Service handles commands from the user
type Service struct {
	tts *tts.Service
}

//Handle processes the incoming command and transforms it into a response
func (s *Service) Handle(wait chan struct{}, userCommands <-chan string) {
	for userCommand := range userCommands {
		//goland:noinspection GoShadowedVar
		userCommand := strings.ToLower(userCommand)
		log.Printf("got command: %q\n", userCommand)
		ctx := context.Background()
		intent := intents.ConvertToIntent(ctx, userCommand)
		intent.Execute(ctx, s.tts)
	}

	close(wait)
}

//New creates a new Service to handle the user commands
func New(ttsService *tts.Service) *Service {
	return &Service{
		tts: ttsService,
	}
}
