package sms

import (
	"go.bug.st/serial"
)

//Service holds all data needed to send SMS
type Service struct {
	port serial.Port
	stub bool
}

//New creates a new SMS gateway
func New(comPort string, baudRate int) (*Service, error) {
	if comPort == "stub" {
		return &Service{port: nil, stub: true}, nil
	}

	mode := &serial.Mode{BaudRate: baudRate}
	port, err := serial.Open(comPort, mode)
	if err != nil {
		return nil, err
	}
	s := &Service{
		port: port,
	}

	return s, nil
}

func (m *Service) send(command string) error {
	err := m.port.ResetOutputBuffer()
	if err != nil {
		return err
	}
	_, err = m.port.Write([]byte(command))
	return err
}

func (m *Service) read(n int) (string, error) {
	var output = ""
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		c, _ := m.port.Read(buf)
		if c > 0 {
			output = string(buf[:c])
		}
	}

	return output, nil
}

//SendSMS sends a message
func (m *Service) SendSMS(phoneNumber, message string) error {
	if m.stub {
		return nil
	}
	if err := m.send("ATE0\r\n"); err != nil {
		return err
	}
	if err := m.send("AT+CMGF=1\r\n"); err != nil {
		return err
	}

	err := m.send("AT+CMGS=\"" + phoneNumber + "\"\r")
	if err != nil {
		return err
	}

	//goland:noinspection GoVetIntToStringConversion
	return m.send(message+string(26))
}
