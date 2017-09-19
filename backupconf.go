/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	ssh "golang.org/x/crypto/ssh"
)

type SSHClient struct {
	ip       string
	user     string
	password string
	client   *ssh.Client
	session  *ssh.Session
	w        io.WriteCloser
	r        io.Reader
}

func (c *SSHClient) Enter(input string) {
	c.w.Write([]byte(input + "\n"))
}

//Expect reads the session stdout and checks if the expected string exists or not.
func (c *SSHClient) Expect(expected string, timeout time.Duration) error {
	buf := bytes.NewBuffer([]byte{})

	rbuf := make([]byte, 32*1024)

	for {
		n, err := c.r.Read(rbuf)
		if err != nil {
			return err
		}

		buf.Write(rbuf[:n])
		fmt.Printf("%s", string(rbuf[:n]))
		if strings.Contains(buf.String(), expected) {
			return nil
		}
	}
}

func (c *SSHClient) Close() {

	if c.client != nil {
		defer c.client.Close()
	}
	if c.session != nil {
		defer c.session.Close()
	}

}

func NewSSHClient(ip string, user string, password string) (*SSHClient, error) {

	sshClient := &SSHClient{ip: ip, user: user, password: password}

	config := &ssh.ClientConfig{
		User: sshClient.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshClient.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc"},
		},
	}
	// config.Config.Ciphers = append(config.Config.Ciphers, "aes128-cbc")
	var err error
	if sshClient.client, err = ssh.Dial("tcp", sshClient.ip+":22", config); err != nil {
		return nil, err
	}
	if sshClient.session, err = sshClient.client.NewSession(); err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sshClient.session.RequestPty("vt100", 80, 40, modes); err != nil {
		return nil, err
	}

	if sshClient.w, err = sshClient.session.StdinPipe(); err != nil {
		return nil, err
	}
	if sshClient.r, err = sshClient.session.StdoutPipe(); err != nil {
		return nil, err
	}

	if _, err = sshClient.session.StderrPipe(); err != nil {
		return nil, err
	}

	if err := sshClient.session.Shell(); err != nil {
		return nil, err
	}

	return sshClient, nil
}

var password = "hi!apple"

func main() {
	c, err := NewSSHClient("10.0.254.151", "back", password)
	checkError(err, "Failed to create SSH client")
	defer c.Close()

	c.Expect(">", 5*time.Second)
	c.Enter("en 5")

	c.Expect("Password:", 5*time.Second)
	c.Enter(password)

	c.Expect("#", 5*time.Second)
	c.Enter("copy run scp")

	c.Expect("Destination filename [scp]?", 5*time.Second)
	c.Enter("cloud@10.99.70.34")

	c.Expect("over write?", 5*time.Second)
	c.Enter("")
	c.Enter("exit")

	//Expect AND, OR?
}

func checkError(err error, info string) {
	if err != nil {
		fmt.Printf("%s. error: %s\n", info, err)
		os.Exit(1)
	}
}
