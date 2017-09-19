/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"bytes"
	"errors"
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
	in       chan int
	out      chan []byte
	quit     chan int
}

func (c *SSHClient) Enter(input string) error {
	_, err := c.w.Write([]byte(input + "\n"))
	return err
}

//Expect reads the session stdout and checks if the expected string exists or not.
func (c *SSHClient) Expect(expected string, timeout time.Duration) error {
	buf := bytes.NewBuffer([]byte{})

	t1 := time.NewTimer(timeout)
	c.in <- 1

	for {
		select {
		case <-t1.C:
			return errors.New("Timeout")
		case res := <-c.out:
			t1.Stop()
			buf.Write(res)
			fmt.Printf("%s", string(res))
			if strings.Contains(buf.String(), expected) {
				return nil
			}
			c.in <- 1
			t1.Reset(timeout)
		}

	}
}

func (c *SSHClient) doRead() error {

	for {
		select {
		case <-c.in:
			rbuf := make([]byte, 32*1024)
			n, err := c.r.Read(rbuf)
			if err != nil {
				fmt.Printf("doRead() exits due to %s\n", err.Error())
				return err
			}
			c.out <- rbuf[:n]
		case <-c.quit:
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

	c.quit <- 1 //doRead() quit

	close(c.in)
	close(c.out)
	close(c.quit)
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

	sshClient.in = make(chan int, 2)
	sshClient.out = make(chan []byte, 2)
	sshClient.quit = make(chan int, 1)

	go sshClient.doRead()

	return sshClient, nil
}

//Changes it to actual password before running
var password = "hi!apple"

func main() {
	c, err := NewSSHClient("10.0.254.151", "back", password)
	checkError(err)
	defer c.Close()

	checkError(c.Expect(">", 5*time.Second))
	checkError(c.Enter("en 5"))

	checkError(c.Expect("Password:", 5*time.Second))
	checkError(c.Enter(password))

	checkError(c.Expect("#", 5*time.Second))
	checkError(c.Enter("copy run scp"))

	checkError(c.Expect("Destination filename [scp]?", 5*time.Second))
	checkError(c.Enter("cloud@10.99.70.34"))

	checkError(c.Expect("over write?", 5*time.Second))
	checkError(c.Enter(""))
	checkError(c.Enter("exit"))

	//Expect AND, OR? How to do

}

func checkError(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
