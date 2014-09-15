package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("repository and commit hash is not given")
	}
	g := &gitCmd{
		repository: os.Args[1],
		branch:     "master",
		commit:     os.Args[2],
		dir:        filepath.Join(os.TempDir(), RandomString(9, 11)),
	}
	if err := g.Clone(); err != nil {
		log.Fatalf("error in cloning %s\n", err)
	}
	fmt.Printf("clone %s in dir %s\n", g.repository, g.dir)
	if err := g.CheckOut(); err != nil {
		log.Fatalf("error in checkout %s\n", err)
	}
	fmt.Printf("checkout commit %s\n", g.commit)
}

type gitCmd struct {
	repository string
	dir        string
	branch     string
	commit     string
}

func lookup(cmd string) (string, error) {
	p, err := exec.LookPath(cmd)
	if err != nil {
		return "", err
	}
	return p, nil
}
func (g *gitCmd) Clone() error {
	path, err := lookup("git")
	if err != nil {
		return err
	}
	cmd := exec.Command(path, "clone", "--branch="+g.branch, g.repository, g.dir)
	var out bytes.Buffer
	var eout bytes.Buffer
	cmd.Stderr = &eout
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, eout.String())
	}
	fmt.Println(out.String())
	return nil
}

func (g *gitCmd) CheckOut() error {
	err := os.Chdir(g.dir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "checkout", "-f", g.commit)
	var out2 bytes.Buffer
	var eout2 bytes.Buffer
	cmd.Stderr = &eout2
	cmd.Stdout = &out2
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("cannot clone\n%s\n%s\n", err, eout2.String())
	}
	return nil
}

// Generates a random string between a range(min and max) of length
func RandomString(min, max int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UTC().UnixNano())
	size := min + rand.Intn(max-min)
	b := make([]byte, size)
	alen := len(alphanum)
	for i := 0; i < size; i++ {
		pos := rand.Intn(alen)
		b[i] = alphanum[pos]
	}
	return string(b)
}
