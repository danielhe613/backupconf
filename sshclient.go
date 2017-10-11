/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	ssh "golang.org/x/crypto/ssh"
)

//SSHClient encapsulates the SSH channel used to communicate with the network device.
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

//Send will write the given string to ssh channel.
func (c *SSHClient) Send(input string) error {
	log.WithFields(log.Fields{
		"target": c.ip,
	}).Infof("Write: %s", input)
	_, err := c.w.Write([]byte(input))
	return err
}

//Expect reads the session stdout and checks if the expected string exists or not.
func (c *SSHClient) Expect(expected string, timeout time.Duration) error {
	buf := bytes.NewBuffer([]byte{})

	t1 := time.NewTimer(timeout)
	c.in <- 1
	time.Sleep(time.Second * 1)

	for {
		select {
		case <-t1.C:
			return errors.New("Read expected string timeout")
		case res := <-c.out:
			buf.Write(res)
			log.WithFields(log.Fields{"target": c.ip}).Infof("Expected: %s, Already Read: %s", expected, buf.String())
			if strings.Contains(buf.String(), expected) {
				t1.Stop()
				return nil
			}
			c.in <- 1
			time.Sleep(time.Second * 1)
			// t1.Reset(timeout)
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
				log.WithFields(log.Fields{
					"target": c.ip,
				}).Infof("doRead() exits due to %s", err.Error())
				return err
			}
			c.out <- rbuf[:n]
		case <-c.quit:
			return nil
		}
	}
}

//Close should be called to release the resources used by SSH client instance.
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

//NewSSHClient is used to create a new SSHClient instance.
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
