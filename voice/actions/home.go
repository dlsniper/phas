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

package actions

import (
	"context"
	"errors"

	"phas/voice/tts"
)

//SayHello will say hello to our users
func SayHello(ctx context.Context, ttsService *tts.Service) error {
	ttsService.Speak(ctx, "Hello Minsk! How are you today?")
	return nil
}

//SetLightsState will set the lights state depending on the user preference
func SetLightsState(ctx context.Context, ttsService *tts.Service) error {
	lightState := ctx.Value("phasLightsState")
	myLightState, ok := lightState.(int)
	if !ok {
		ttsService.Speak(ctx, "I could not change the lights state.")
		return errors.New("failed to set lights state")
	}

	//TODO Undo changes for demo
	//TODO Don't play with the lights when on stage, someone is at home
	_ = myLightState

	return nil
}
