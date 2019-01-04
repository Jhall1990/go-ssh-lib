package sshlib

import (
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

/*
ErrNoConnection - Error returned when the agent is unable to establish a connection.
*/
var ErrNoConnection = errors.New("unable to establish connection")

/*
ErrNoMatch - Error returned when no match is found in the returned data.
*/
var ErrNoMatch = errors.New("unable to find match")

/*
ErrLostConnection - Error returned when when an established connection is lost.
*/
var ErrLostConnection = errors.New("lost connection and unable to re-establish")

/*
ErrNoPrompt - Error returned when the prompt could not be located after logging in.
*/
var ErrNoPrompt = errors.New("unable to locate prompt")

/*
ErrInvalidAgent - Error returned when the agent string passed is not supported.
*/
var ErrInvalidAgent = errors.New("invalid agent type")

/*
SSHLib - A library for SSH.
*/
type SSHLib struct {
	Host      string
	Port      string
	Conn      *ssh.Session
	Buffer    string
	User      string
	Passwd    string
	PromptReg string
	Stdin     io.WriteCloser
	Stdout    io.Reader
	Data      chan string
	timeout   time.Duration
}

/*
CreateSSH - Creates an instance of the SSHLib struct.
*/
func CreateSSH(host, port, user, passwd, promptReg string, timeout int) SSHLib {
	s := SSHLib{}
	s.Host = host
	s.Port = port
	s.User = user
	s.Passwd = passwd
	s.PromptReg = promptReg
	s.Data = make(chan string)
	s.timeout = time.Duration(timeout) * time.Second

	return s
}

/*
Open - Opens an ssh connection.
*/
func (s *SSHLib) Open() error {
	sshConfig := s.CreateSSHConfig()
	conn, err := ssh.Dial("tcp", s.Host+":"+s.Port, sshConfig)

	if err != nil {
		return ErrNoConnection
	}

	s.Conn, err = conn.NewSession()

	if err != nil {
		return ErrNoConnection
	}

	s.Stdin, _ = s.Conn.StdinPipe()
	s.Stdout, _ = s.Conn.StdoutPipe()
	modes := ssh.TerminalModes{ssh.ECHO: 0}
	s.Conn.RequestPty("vt220", 40, 500, modes)
	s.Conn.Shell()
	go s.reader()

	_, err = s.ReadUntilRegex(s.PromptReg, 3)

	if err != nil {
		return ErrNoPrompt
	}

	return nil
}

/*
CreateSSHConfig - Creates the SSH configuration object.
*/
func (s *SSHLib) CreateSSHConfig() *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		Timeout:         s.timeout,
		User:            s.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Passwd),
		},
	}
	return config
}

/*
ReadUntil - ReadUntil - Reads until the string s is seen. Then returns the output.
			If not match is found in <timeout> seconds, return whatever
			is in the buffer.
*/
func (s *SSHLib) ReadUntil(str string, timeout int) (string, error) {
	var returnData string

	foundMatch := false
	startTime := getTimestamp()
	timeoutMs := int64(timeout) * 1000

	for getTimestamp()-startTime < timeoutMs {
		if strings.Contains(s.Buffer, str) {
			var endRead = strings.Index(s.Buffer, str)
			returnData = s.Buffer[:endRead+len(str)]
			s.Buffer = s.Buffer[endRead:]
			foundMatch = true
			break
		} else {
			s.GetData()
		}
	}

	if returnData == "" {
		returnData = s.Buffer
		s.Buffer = ""
	}

	if !foundMatch {
		return returnData, ErrNoMatch
	}
	return returnData, nil
}

/*
ReadUntilRegex - Read until the a regex match is found. Then return the output.
				 If not match is found within <timeout> seconds, return whatever
				 is in the buffer.
*/
func (s *SSHLib) ReadUntilRegex(regexStr string, timeout int) (string, error) {
	var returnData string

	foundMatch := false
	startTime := getTimestamp()
	timeoutMs := int64(timeout) * 1000
	cmpRe, err := regexp.Compile(regexStr)

	if err != nil {
		return "", err
	}

	for getTimestamp()-startTime < timeoutMs {
		if cmpRe.MatchString(s.Buffer) == true {
			var endRead = cmpRe.FindStringIndex(s.Buffer)[1]
			returnData = s.Buffer[:endRead]
			s.Buffer = s.Buffer[endRead:]
			foundMatch = true
			break
		} else {
			s.GetData()
		}
	}

	if returnData == "" {
		returnData = s.Buffer
		s.Buffer = ""
	}

	if !foundMatch {
		return returnData, ErrNoMatch
	}
	return returnData, nil
}

/*
ReadUntilRegexList - Read until a match is found for one of the regex strings in the list.
					 Then return the output. If not match is found within <timeout> seconds,
					 return whatever is in the buffer.
*/
func (s *SSHLib) ReadUntilRegexList(regexList []string, timeout int) (string, error) {
	var returnData string
	var cmpReList []*regexp.Regexp

	foundMatch := false
	startTime := getTimestamp()
	timeoutMs := int64(timeout) * 1000

	for i := 0; i < len(regexList); i++ {
		var cmpRe, _ = regexp.Compile(regexList[i])
		cmpReList = append(cmpReList, cmpRe)
	}

	for !foundMatch && getTimestamp()-startTime < timeoutMs {
		for i := 0; i < len(cmpReList); i++ {
			if cmpReList[i].MatchString(s.Buffer) == true {
				var endRead = cmpReList[i].FindStringIndex(s.Buffer)[1]
				returnData = s.Buffer[:endRead]
				s.Buffer = s.Buffer[endRead:]
				foundMatch = true
			}
		}
		if foundMatch == false {
			s.GetData()
		}
	}

	if returnData == "" {
		returnData = s.Buffer
		s.Buffer = ""
	}

	if !foundMatch {
		return returnData, ErrNoMatch
	}
	return returnData, nil
}

/*
GetData - Attempts to pull data from s.Data channel, if nothing is present sleep for 250ms.
*/
func (s *SSHLib) GetData() {
	select {
	case data := <-s.Data:
		s.Buffer += data
	default:
		time.Sleep(250 * time.Millisecond)
	}
}

/*
reader - Reads data from stdout in a loop and adds it to s.Data channel.
*/
func (s *SSHLib) reader() {
	recvData := make([]byte, 1024)
	for {
		numBytes, _ := s.Stdout.Read(recvData)
		s.Data <- string(recvData[:numBytes])
	}
}

/*
getTimestamp - Gets the current number of milliseconds since epoch.
*/
func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

/*
Write - Writes the string plus a new line to the socket.
*/
func (s *SSHLib) Write(str string) {
	stringBytes := []byte(str + "\n")
	s.Stdin.Write(stringBytes)
}

/*
WriteThenReadUntil - Writes a string plus a new line then waits for the string s to be present.
                     Then returns the output.
*/
func (s *SSHLib) WriteThenReadUntil(sendStr string, matchStr string, timeout int) (string, error) {
	s.Write(sendStr)
	return s.ReadUntil(matchStr, timeout)
}

/*
Close - Closes the telnet socket.
*/
func (s *SSHLib) Close() {
	s.Conn.Close()
}
