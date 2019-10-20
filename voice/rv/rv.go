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

package rv

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gordonklaus/portaudio"
)

//Service that handles the voice recording
type Service struct{}

func (s *Service) newTempFile() *os.File {
	tempFile, err := ioutil.TempFile("", "command-*.aiff")
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = tempFile.WriteString("FORM"); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int32(0)); err != nil {
		log.Fatalln(err)
	}
	if _, err = tempFile.WriteString("AIFF"); err != nil {
		log.Fatalln(err)
	}

	if _, err = tempFile.WriteString("COMM"); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int32(18)); err != nil {
		log.Fatalln(err)
	}

	if err = binary.Write(tempFile, binary.BigEndian, int16(1)); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int32(0)); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int16(16)); err != nil {
		log.Fatalln(err)
	}
	if _, err = tempFile.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}); err != nil {
		log.Fatalln(err)
	}

	if _, err = tempFile.WriteString("SSND"); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int32(0)); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int32(0)); err != nil {
		log.Fatalln(err)
	}
	if err = binary.Write(tempFile, binary.BigEndian, int32(0)); err != nil {
		log.Fatalln(err)
	}

	return tempFile
}

func (s *Service) closeFile(tempFile *os.File, nSamples int) {
	totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
	if _, err := tempFile.Seek(4, 0); err != nil {
		log.Fatalln(err)
	}
	if err := binary.Write(tempFile, binary.BigEndian, int32(totalBytes)); err != nil {
		log.Fatalln(err)
	}
	if _, err := tempFile.Seek(22, 0); err != nil {
		log.Fatalln(err)
	}
	if err := binary.Write(tempFile, binary.BigEndian, int32(nSamples)); err != nil {
		log.Fatalln(err)
	}
	if _, err := tempFile.Seek(42, 0); err != nil {
		log.Fatalln(err)
	}
	if err := binary.Write(tempFile, binary.BigEndian, int32(4*nSamples+8)); err != nil {
		log.Fatalln(err)
	}
	if err := tempFile.Close(); err != nil {
		log.Fatalln(err)
	}
}

func (s *Service) removeFile(filename string) {
	if err := os.Remove(filename); err != nil {
		log.Println(err)
	}
}

func (s *Service) recordVoice() string {
	nSamples, tempFile := 0, s.newTempFile()

	// prevent Go from capturing the value of nSamples at the beginning of the function, rendering the call useless
	defer func() { s.closeFile(tempFile, nSamples) }()

	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 16000, len(in), in)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	if err := stream.Start(); err != nil {
		log.Fatalln(err)
	}

	log.Println("listening for command...")
	func() {
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for {
			if err := stream.Read(); err != nil {
				log.Fatalln(err)
			}
			if err := binary.Write(tempFile, binary.BigEndian, in); err != nil {
				log.Fatalln(err)
			}
			nSamples += len(in)
			select {
			case <-timer.C:
				return
			default:
			}
		}
	}()
	if err := stream.Stop(); err != nil {
		log.Fatalln(err)
	}
	log.Println("finished listening for command...")

	return tempFile.Name()
}

func (s *Service) processVoice(voiceFile string) []byte {
	tempFile, err := ioutil.TempFile("", "converted-speech-*.wav")
	if err != nil {
		log.Fatalln(err)
	}
	defer s.removeFile(tempFile.Name())
	defer func() {
		err := tempFile.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	vFile, err := ioutil.ReadFile(voiceFile)
	if err != nil {
		log.Fatalln(err)
	}
	decoded := aiff.NewDecoder(bytes.NewReader(vFile))
	if !decoded.IsValidFile() {
		log.Fatalln("voice file is not a valid aiff file")
	}

	enc := wav.NewEncoder(tempFile, decoded.SampleRate, int(decoded.BitDepth), int(decoded.NumChans), 1)
	format := &audio.Format{
		NumChannels: int(decoded.NumChans),
		SampleRate:  decoded.SampleRate,
	}

	bufferSize := 1000000
	buf := &audio.IntBuffer{Data: make([]int, bufferSize), Format: format, SourceBitDepth: int(decoded.BitDepth)}

	var n int
	for {
		n, err = decoded.PCMBuffer(buf)
		if err != nil {
			break
		}

		if n == 0 {
			break
		}

		if n != len(buf.Data) {
			buf.Data = buf.Data[:n]
		}

		if err := enc.Write(buf); err != nil {
			log.Fatalln(err)
		}
	}

	if err := enc.Close(); err != nil {
		log.Fatalln(err)
	}

	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		log.Fatalln(err)
	}

	return content
}

//Listen begins listening for user input
func (s *Service) Listen() []byte {
	fileName := s.recordVoice()

	defer s.removeFile(fileName)

	return s.processVoice(fileName)
}

//NewService creates a new voice recording service
func NewService() *Service {
	return &Service{}
}
