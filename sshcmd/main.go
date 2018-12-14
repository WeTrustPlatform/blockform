package sshcmd

import (
	"time"

	"golang.org/x/crypto/ssh"
)

// Exec executes a simple command over SSH and returns an error if any
func Exec(privKey, passphrase, user, address, cmd string) error {
	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(privKey), []byte(passphrase))
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", address+":22", config)
	if err != nil {
		return err
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// sess.Stdout = os.Stdout
	// sess.Stderr = os.Stderr

	err = sess.Run(cmd)
	if err != nil {
		return err
	}

	return nil
}
