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

package hue

import (
	"fmt"
	"time"

	"github.com/amimof/huego"
)

//Service holds all the lightning service data
type Service struct {
	bridge *huego.Bridge
}

//New creates a new hue Service
func New(address, user string) *Service {
	res := &Service{}
	if address == "" || user == "" {
		return res
	}
	res.bridge = huego.New(address, user)
	return res
}

//Alarm triggers the alarm
func (s *Service) Alarm(groupName string) error {
	if s.bridge == nil {
		return nil
	}
	gs, err := s.bridge.GetGroups()
	if err != nil {
		return err
	}

	gr := huego.Group{}
	for _, g := range gs {
		if g.Name == groupName {
			gr = g
			break
		}
	}
	if gr.Name != groupName {
		return fmt.Errorf("group not found")
	}

	iState := *gr.State

	state := huego.State{
		On:    true,
		Alert: "lselect",
		Bri:   254,
		Sat:   254,
		Hue:   65535,
	}
	err = gr.SetState(state)
	if err != nil {
		return err
	}

	time.Sleep(16 * time.Second)

	// Reset the alarm state and prevent the color mode from erroring out
	iState.Alert = "none"
	iState.ColorMode = ""
	return gr.SetState(iState)
}
