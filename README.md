# go-ssh-lib

Simple go ssh client that utilizes the go ssh package. Supports sending multiple commands per session.

## Getting Started

The simplest way to use this is to use the provided SSH agent, it provides convience methods for seding commands via SSH.

```
s, err := sshlib.Create(<ip>, <port>, <username>, <password>, <prompt_regex>, <timeout>)

if err != nil {
	fmt.Println(err.Error())
}

out, err := s.SendCommand("ls")

if err != nil {
	fmt.Println(err.Error())
}

fmt.Println(out)
```

## Agent Methods

Create - Creates an instance of the SSH agent and set the needed attributes.

Connect - Attempts to connect to the given device.

Logout - Logs out of the given device.

SendCommandNoWait - Sends a command and does not wait for any return value.

SendCommand - Sends a command waits for promptRegex before returning.

SendCommandStripCommand - Sends a command and removes the command from the beginning of the output.

SendCommandWaitForList - Sends a command and waits for one of the regex strings in the list.
                         The prompt regex is added automatically.

SetConnected - Sets the ssh agent structs connected flag.

GetConnected - Returns the ssh agent structs connected flag.

