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
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dlsniper/phas/wakeword"
)

func initializeWakeWordListener() *wakeword.Listener {
	wd, _ := os.Getwd()
	// export PHAS_OS=mac
	// export PHAS_OS=linux
	// export PHAS_OS=windows
	// export PHAS_OS=raspberrypi

	runningOS := strings.ToLower(os.Getenv("PHAS_OS"))
	if runningOS == "" {
		runningOS = runtime.GOOS
	}
	modelPath := wd + "/lib/common/porcupine_params.pv"
	libDir := fmt.Sprintf("%s/lib/resources/%s", wd, runningOS)
	keywords := []wakeword.Keyword{
		{
			Name:        "bumblebee",
			FilePath:    fmt.Sprintf("%s/bumblebee_%s.ppn", libDir, runningOS),
			Sensitivity: 0.7,
		},
		{
			Name:        "terminator",
			FilePath:    fmt.Sprintf("%s/terminator_%s.ppn", libDir, runningOS),
			Sensitivity: 0.8,
		},
	}

	return wakeword.NewListener(modelPath, keywords)
}
