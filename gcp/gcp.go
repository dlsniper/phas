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

package gcp

import (
	"context"
	"log"

	speech "cloud.google.com/go/speech/apiv1"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
)

func authenticate(ctx context.Context) (*cloudkms.Service, error) {
	oauthClient, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	opts := option.WithHTTPClient(oauthClient)
	return cloudkms.NewService(ctx, opts)
}

//InitServices Initializes the GCP services needed by the application
func InitServices(ctx context.Context) (*speech.Client, *texttospeech.Client) {
	_, err := authenticate(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	stt, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	tts, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return stt, tts
}
