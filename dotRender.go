package main

import (
	"github.com/pkg/errors"
	"log"
	"os/exec"
)

func renderPng(dotData []byte, path string) error {
	cmd := exec.Command("dot", "-Tpng", "-Gdpi=50", "-o", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "Failed getting STDIN pipe for DOT command")
	}

	go func() {
		defer stdin.Close()
		stdin.Write(dotData)
	}()
	out, err := cmd.CombinedOutput()

	log.Printf("%s\n", out)

	if err != nil {
		return errors.Wrap(err, "Failed running DOT command")
	}
	return nil
}
