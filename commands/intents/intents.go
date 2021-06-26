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

package intents

import (
	"context"
	"log"

	"github.com/dlsniper/phas/tts"
)

//An Action runs whenever a user command matches with an Intent
type Action func(context.Context, *tts.Service) error

//An Intent contains of a command, alternative ways to give the command, and a series of actions that must run
type Intent struct {
	Command      string
	Alternatives []string
	Actions      []Action
}

// Weather in [city]

//Matches method checks if an Intent matches a given command
func (i *Intent) Matches(_ context.Context, command string) bool {
	// TODO Add placeholder support and return the found pairings as well
	if command == i.Command {
		return true
	}

	for _, cmd := range i.Alternatives {
		if cmd == command {
			return true
		}
	}

	return false
}

//Execute runs the given actions for the current Intent
func (i *Intent) Execute(ctx context.Context, tts *tts.Service) {
	for idx, action := range i.Actions {
		err := action(ctx, tts)
		if err != nil {
			log.Printf("error %v while executing the intent %q at action %d\n", err, i.Command, idx)
			// TODO Handle errors that will allow the rest of the intent to run
			// TODO Tell users about the errors encountered while running the intent and ask if the intent should continue the execution or retry
			break
		}
	}
}

var noMatchingIntent = &Intent{
	Command: "",
	Actions: []Action{
		func(ctx context.Context, tts *tts.Service) error {
			tts.Speak(ctx, "I could not understand your request. Please try again.")
			return nil
		},
	},
}

var intents []*Intent

//ConvertToIntent handles converting the given command to an Intent
func ConvertToIntent(ctx context.Context, command string) *Intent {
	for _, intent := range intents {
		if intent.Matches(ctx, command) {
			return intent
		}
	}

	return noMatchingIntent
}

//RegisterIntent registers the available Intents
func RegisterIntent(intent *Intent) {
	intents = append(intents, intent)
}
