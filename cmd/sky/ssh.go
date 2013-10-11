package main

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/gopass"
	"errors"
	"github.com/skynetservices/skynet/log"
)

type SSHConn struct {
	host   string
	user   string
	client *ssh.ClientConn
	env    map[string]string
}

// SSH logic
func (c *SSHConn) Connect(host, user string) error {
	c.host = host
	c.user = user

	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(c),
		},
	}

	var err error
	c.client, err = ssh.Dial("tcp", c.host, config)
	if err != nil {
		return err
	}

	return nil
}

func (c *SSHConn) ExecPath(cmd, path string) (out []byte, err error) {
	var session *ssh.Session

	session, err = c.client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: " + err.Error())
	}
	defer session.Close()

	envVars := ""
	if c.env != nil {
		for name, value := range c.env {
			envVars = envVars + name + "=\"" + value + "\" "
			/*
				      TODO: This should be the proper way to set the environment, but fails for some reason
							       * Investigate why and possibly send pull-request to maintainer
										err = session.Setenv(name, value)

										if err != nil {
											log.Fatal("Failed to set environment: " + err.Error())
										}
			*/
		}
	}

	cmd = envVars + cmd

	if path != "" {
		cmd = "cd " + path + " && " + cmd
	}

	return session.CombinedOutput(cmd)
}

func (c *SSHConn) Exec(cmd string) (out []byte, err error) {
	return c.ExecPath(cmd, "")
}

func (c *SSHConn) Password(user string) (string, error) {
	pass, err := gopass.GetPass("Password for " + user + ": ")

	if err != nil {
		return "", errors.New("Failed to collect password: " + err.Error())
	}

	return pass, err
}

func (c *SSHConn) Close() {
	c.client.Close()
}

func (c *SSHConn) SetEnv(name, value string) {
	if c.env == nil {
		c.env = make(map[string]string, 5)
	}

	c.env[name] = value
}
