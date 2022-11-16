package main

import (
	"github.com/pkg/errors"
	"os/exec"
)

func renderPng(dotData []byte, path string) error {
	cmd := exec.Command("dot", "-Tpng", "-o", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "Failed getting STDIN pipe for DOT command")
	}

	go func() {
		defer stdin.Close()
		stdin.Write(dotData)
	}()

	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "Failed running DOT command")
	}
	return nil
}
