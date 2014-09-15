package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"

	"code.google.com/p/go.crypto/ssh"

	"gopkg.in/codegangsta/cli.v0"
)

func main() {
	app := cli.NewApp()
	app.Name = "scopy"
	app.Usage = "securely copy files to a remote server over ssh"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "keyfile, k",
			Usage: "Full path to a ssh private key file",
		},
		cli.StringFlag{
			Name:  "from, f",
			Usage: "Name of local folder from where file will be copied",
		},
		cli.StringFlag{
			Name:  "to, t",
			Usage: "Name of the remote folder where the file will be copied",
		},
		cli.StringFlag{
			Name:  "host",
			Usage: "Remote host",
		},
		cli.StringFlag{
			Name:  "user, u",
			Usage: "Remote user",
		},
	}
	app.Action = copyToRemote
	app.Run(os.Args)
}

func copyToRemote(c *cli.Context) {
	for _, flag := range []string{"keyfile", "to", "host", "user"} {
		if c.Generic(flag) == nil {
			log.Fatalf("flag %s is required\n", flag)
		}
	}
	// process the keyfile
	buf, err := ioutil.ReadFile(c.String("keyfile"))
	if err != nil {
		log.Fatalf("error in reading private key file %s\n", err)
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		log.Fatalf("error in parsing private key %s\n", key)
	}
	// client config
	config := &ssh.ClientConfig{
		User: c.String("user"),
		Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
	}
	// connection
	client, err := ssh.Dial("tcp", c.String("host")+":22", config)
	if err != nil {
		log.Fatalf("error in ssh connection %s\n", err)
	}
	defer client.Close()
	// sftp handler
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatalf("error in sftp connection %s\n", err)
	}
	defer sftp.Close()
	// Remote file
	r, err := sftp.Create(filepath.Join(c.String("to"), "go_sftp.txt"))
	if err != nil {
		log.Fatalf("error in creating remote file %s\n", err)
	}
	l := strings.NewReader("Writing through golang sftp")
	if _, err := io.Copy(r, l); err != nil {
		log.Fatalf("error in writing file to remote system %s\n", err)
	}
	log.Println("written new file to remote system")
}
