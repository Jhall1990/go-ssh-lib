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

