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

package main

import (
	"context"

	"github.com/dlsniper/phas/actions"
	"github.com/dlsniper/phas/commands/intents"
	"github.com/dlsniper/phas/sentry"
	"github.com/dlsniper/phas/tts"
)

func registerIntents(s *sentry.Service) {
	myIntents := []*intents.Intent{
		{
			Command: "turn the lights on",
			Alternatives: []string{
				"turn on the lights",
			},
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					ctx = context.WithValue(ctx, "phasLightsState", 255)
					return actions.SetLightsState(ctx, ttsService)
				},
			},
		},
		{
			Command: "turn the lights off",
			Alternatives: []string{
				"turn off the lights",
			},
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					ctx = context.WithValue(ctx, "phasLightsState", 0)
					return actions.SetLightsState(ctx, ttsService)
				},
			},
		},
		{
			Command: "dim the lights",
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					ctx = context.WithValue(ctx, "phasLightsState", 70)
					return actions.SetLightsState(ctx, ttsService)
				},
			},
		},
		{
			Command: "turn on the sentry mode",
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					ctx = context.WithValue(ctx, "sentryMode", true)
					return actions.SentryMode(ctx, ttsService, s)
				},
			},
		},
		{
			Command: "turn off the sentry mode",
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					ctx = context.WithValue(ctx, "sentryMode", false)
					return actions.SentryMode(ctx, ttsService, s)
				},
			},
		},
		{
			Command: "tell me a joke",
			Alternatives: []string{
				"tell a joke",
			},
			Actions: []intents.Action{
				actions.TellAJoke,
			},
		},
		{
			Command: "say hello",
			Alternatives: []string{
				"say hi",
			},
			Actions: []intents.Action{
				actions.SayHello,
			},
		},
	}

	for _, intent := range myIntents {
		intents.RegisterIntent(intent)
	}
}
