package sshlib

import (
	"strings"
)

/*
SSHAgent - A basic go telnet agent.
*/
type SSHAgent struct {
	Host        string
	Port        string
	User        string
	Passwd      string
	PromptRegex string
	Timeout     int
	Conn        SSHLib
	Connected   bool
}

/*
Create - Creates an instance of the SSH agent and set the needed attributes.
*/
func Create(host, port, user, passwd, promptRegex string, timeout int) (*SSHAgent, error) {
	a := &SSHAgent{}
	a.Host = host
	a.Port = port
	a.User = user
	a.Passwd = passwd
	a.PromptRegex = promptRegex
	a.Timeout = timeout
	a.Conn = CreateSSH(host, port, user, passwd, promptRegex, timeout)
	err := a.Connect()

	if err != nil {
		return a, err
	}

	return a, nil
}

/*
Connect - Attempts to connect to the given device.
*/
func (s *SSHAgent) Connect() error {
	if s.Connected == false {
		err := s.Conn.Open()

		if err != nil {
			return err
		}
		s.Connected = true
	}
	return nil
}

/*
Login - Logs into the given device.
*/
func (s *SSHAgent) Login() error {
	return nil
}

/*
Logout - Logs out of the given device.
*/
func (s *SSHAgent) Logout() {
	s.Conn.Close()
}

/*
SendCommandNoWait - Sends a command and does not wait for any return value.
*/
func (s *SSHAgent) SendCommandNoWait(command string) error {
	err := s.Connect()

	if err != nil {
		return ErrLostConnection
	}

	s.Conn.Write(command)

	return nil
}

/*
SendCommand - Sends a command waits for promptRegex before returning.
*/
func (s *SSHAgent) SendCommand(command string) (string, error) {
	err := s.SendCommandNoWait(command)

	if err != nil {
		return "", ErrLostConnection
	}

	output, _ := s.Conn.ReadUntilRegex(s.PromptRegex, s.Timeout)

	return output, nil
}

/*
SendCommandStripCommand - Sends a command and removes the command from the beginning of the output.
*/
func (s *SSHAgent) SendCommandStripCommand(command string) (string, error) {
	output, err := s.SendCommand(command)

	if err != nil {
		return "", err
	}

	outputList := strings.Split(output, "\n")

	if strings.Contains(outputList[0], command) {
		return strings.Join(outputList[1:], "\n"), nil
	}
	return output, nil
}

/*
SendCommandWaitForList - Sends a command and waits for one of the regex strings in the list.
                         The prompt regex is added automatically.
*/
func (s *SSHAgent) SendCommandWaitForList(command string, regexList []string) (string, error) {
	err := s.SendCommandNoWait(command)

	if err != nil {
		return "", err
	}
	regexList = append(regexList, s.PromptRegex)
	output, err := s.Conn.ReadUntilRegexList(regexList, s.Timeout)

	if err != nil {
		return output, err
	}

	return output, nil
}

/*
SetConnected - Sets the ssh agent structs connected flag.
*/
func (s *SSHAgent) SetConnected(conFlag bool) {
	s.Connected = conFlag
}

/*
GetConnected - Returns the ssh agent structs connected flag.
*/
func (s *SSHAgent) GetConnected() bool {
	return s.Connected
}
