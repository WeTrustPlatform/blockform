package sshcmd

import (
	"bytes"
	"time"

	"golang.org/x/crypto/ssh"
)

// Exec executes a simple command over SSH and returns an error if any
func Exec(privKey, passphrase, user, address, cmd string) (string, string, error) {
	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(privKey), []byte(passphrase))
	if err != nil {
		return "", "", err
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", address+":22", config)
	if err != nil {
		return "", "", err
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer sess.Close()

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")

	sess.Stdout = stdout
	sess.Stderr = stderr

	err = sess.Run(cmd)
	if err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}
