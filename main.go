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
//
// Name: Florin Pățan
// Work: Developer Advocate @ JetBrains (GoLand IDE)
// Blog: https://blog.jetbrains.com/go
// Twitter: @GoLandIDE
// Twitter: @DLSNIPER
// Mail: florin@jetbrains.com
// Gophers Slack: @DLSNIPER (https://invite.slack.golangbridge.org) #goland



package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"phas/voice/actions"
	"phas/voice/commands"
	"phas/voice/commands/intents"
	"phas/voice/gcp"
	"phas/voice/rv"
	"phas/voice/stt"
	"phas/voice/tts"
	"phas/voice/wakeword"

	"github.com/charithe/porcupine-go"
	"github.com/gordonklaus/portaudio"
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx := context.Background()

	wait := make(chan struct{})
	userCommands := make(chan string, 10)

	if err := portaudio.Initialize(); err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := portaudio.Terminate(); err != nil {
			log.Fatalln(err)
		}
	}()

	wwListener := initializeWakeWordListener()

	sttClient, ttsClient := gcp.InitServices(ctx)
	sttService := stt.NewService(sttClient)
	ttsService := tts.NewService(ttsClient)
	commandListener := rv.NewService()
	commandsService := commands.NewService(ttsService)

	registerIntents()

	go commandsService.Handle(wait, userCommands)

	for {
		word, cx := wwListener.Listen(ctx)
		if word == "terminator" {
			break
		}

		userCommands <- sttService.Process(cx, commandListener.Listen())
	}

	//Clean shutdown of the system
	close(userCommands)
	<-wait
	ttsService.Speak(ctx, "I'll be back, Florence!")
	log.Println("received exit command")
}

func initializeWakeWordListener() *wakeword.Listener {
	wd, _ := os.Getwd()
	runningOS := strings.ToLower(os.Getenv("PHAS_OS"))
	modelPath := wd + "/lib/common/porcupine_params.pv"
	libDir :=  fmt.Sprintf("%s/lib/resources/%s", wd, runningOS)
	keywords := []*porcupine.Keyword{
		{
			Value:       "bumblebee",
			FilePath:    fmt.Sprintf("%s/bumblebee_%s.ppn", libDir, runningOS),
			Sensitivity: 0.7,
		},
		{
			Value:       "terminator",
			FilePath:    fmt.Sprintf("%s/terminator_%s.ppn", libDir, runningOS),
			Sensitivity: 0.8,
		},
	}

	return wakeword.NewListener(modelPath, keywords)
}

func registerIntents() {
	var myIntents = []*intents.Intent{
		{
			Command: "turn the lights on",
			Alternatives: []string{
				"turn on the lights",
			},
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					ctx = context.WithValue(ctx, "allLightsState", 255)
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
					ctx = context.WithValue(ctx, "allLightsState", 0)
					return actions.SetLightsState(ctx, ttsService)
				},
			},
		},
		{
			Command: "say hello",
			Alternatives: []string{
				"say hi",
			},
			Actions: []intents.Action{
				func(ctx context.Context, ttsService *tts.Service) error {
					return actions.SayHello(ctx, ttsService)
				},
			},
		},
		//TODO Restore after demo
	}

	for _, intent := range myIntents {
		intents.RegisterIntent(intent)
	}
}
