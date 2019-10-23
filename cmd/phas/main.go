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
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/dlsniper/phas/commands"
	"github.com/dlsniper/phas/gcp"
	"github.com/dlsniper/phas/hue"
	"github.com/dlsniper/phas/rv"
	"github.com/dlsniper/phas/sentry"
	"github.com/dlsniper/phas/sms"
	"github.com/dlsniper/phas/stt"
	"github.com/dlsniper/phas/tts"
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx := context.Background()

	wait := make(chan struct{})
	userCommands := make(chan string, 10)

	wwListener := initializeWakeWordListener()

	sttClient, ttsClient := gcp.InitServices(ctx)
	sttService := stt.New(sttClient)
	ttsService := tts.New(ttsClient)
	commandListener := rv.New()
	commandsService := commands.New(ttsService)

	smsCOMPort := os.Getenv("PHAS_SMS_COM_PORT")
	if smsCOMPort == "" {
		smsCOMPort = "stub"
	}
	smsService, err := sms.New(smsCOMPort, 115200)
	if err != nil {
		log.Fatalln(err)
	}

	hueAddr, hueUser := os.Getenv("PHAS_HUE_ADDR"), os.Getenv("PHAS_HUE_USER")
	lightsService := hue.New(hueAddr, hueUser)

	cameraID := 0
	cam := os.Getenv("PHAS_SENTRY_CAM")
	if cam == "" && runtime.GOOS == "windows" {
		// Take a shortcut in development
		cameraID = 1
	} else if cam != "" {
		var err error
		cameraID, err = strconv.Atoi(cam)
		if err != nil {
			cameraID = 0
		}
	}

	sentryLightGroup := os.Getenv("PHAS_SENTRY_LIGHT_GROUP")
	sentryPhoneNumber := os.Getenv("PHAS_SENTRY_PHONE")

	sentryService := sentry.New(cameraID, 3000, wait, func(sinceLastAlarm float64) {
		if sinceLastAlarm > 20 {
			ttsService.Speak(ctx, "Intruder detected! Sound the alarm!")

			err := smsService.SendSMS(sentryPhoneNumber, "Intruder detected! Sound the alarm!")
			if err != nil {
				log.Println(err)
			}

			err = lightsService.Alarm(sentryLightGroup)
			if err != nil {
				log.Println(err)
			}
		}
	})

	registerIntents(sentryService)

	// Handle sends a close message when done
	go commandsService.Handle(wait, userCommands)

	for {
		log.Println("waiting for wakewords")
		word, cx := wwListener.Listen(ctx)
		if word == "terminator" {
			break
		}
		userCommands <- sttService.Process(cx, commandListener.Listen())
	}

	//Clean shutdown of the system
	close(userCommands)
	<-wait
	ttsService.Speak(ctx, "I'll be back!")
	log.Println("got command: exit")
}
