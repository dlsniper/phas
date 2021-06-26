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

package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dlsniper/phas/sentry"
	"github.com/dlsniper/phas/tts"
)

//SayHello will say hello to our users
func SayHello(ctx context.Context, ttsService *tts.Service) error {
	ttsService.Speak(ctx, "Hello, Human! How are you today?")
	return nil
}

//SetLightsState will set the hue state depending on the user preference
func SetLightsState(ctx context.Context, ttsService *tts.Service) error {
	lightState := ctx.Value("phasLightsState")
	myLightState, ok := lightState.(int)
	if !ok {
		ttsService.Speak(ctx, "I could not change the hue state.")
		return errors.New("failed to set hue state")
	}

	ttsService.Speak(ctx, fmt.Sprintf("Changing the hue to state %d.", myLightState))

	return nil
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		ExpectContinueTimeout: 3 * time.Second,
	},
}

//TellAJoke for the audience
func TellAJoke(ctx context.Context, ttsService *tts.Service) error {
	url := "https://official-joke-api.appspot.com/random_joke"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	r, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	type joke struct {
		Setup string `json:"setup,omitempty"`
		Punchline string `json:"punchline,omitempty"`
	}

	j := joke{}
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	ttsService.Speak(ctx, j.Setup+"    "+j.Punchline)
	return nil
}

//SentryMode will set the hue state depending on the user preference
func SentryMode(ctx context.Context, ttsService *tts.Service, s *sentry.Service) error {
	sentryModeState := ctx.Value("sentryMode")
	desiredSentryMode, ok := sentryModeState.(bool)
	if !ok {
		ttsService.Speak(ctx, "I could not set the sentry mode state.")
		return errors.New("failed to sentry mode state")
	}

	s.Toggle(ctx, ttsService, desiredSentryMode)

	return nil
}
