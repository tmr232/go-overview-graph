package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
)

//go:embed side-by-side.html.tmpl
var sideBySideTemplate string

type sideBySideRow struct {
	Code  string
	Image *string
}

func (row sideBySideRow) HasImage() bool {
	return row.Image != nil
}

type sideBySideData struct {
	Filename string
	Rows     []sideBySideRow
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open "+path)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "Error while reading "+path)
	}
	return lines, nil
}

func renderSideBySide(fileOverview *FileOverview, outPath string) error {
	lines, err := readLines(fileOverview.Filename)
	if err != nil {
		return errors.Wrap(err, "Failed reading lines of file "+fileOverview.Filename)
	}

	functionOverviews := fileOverview.Functions
	sort.Slice(functionOverviews, func(i, j int) bool {
		return functionOverviews[i].Line < functionOverviews[j].Line
	})

	data := sideBySideData{
		Filename: fileOverview.Filename,
		Rows:     nil,
	}
	var lineBlocks []string
	currentLine := 0

	for _, overview := range functionOverviews {
		lineBlocks = append(lineBlocks, strings.Join(lines[currentLine:overview.Line], "\n"))
		currentLine = overview.Line
	}
	lineBlocks = append(lineBlocks, strings.Join(lines[currentLine:], "\n"))

	data.Rows = append(data.Rows, sideBySideRow{
		Code:  lineBlocks[0],
		Image: nil,
	})

	for i, code := range lineBlocks[1:] {
		funcOverview := functionOverviews[i]

		imagePath := path.Join(outPath, funcOverview.Name+".png")
		err := renderPng([]byte(funcOverview.Dot), imagePath)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed rendering PNG for %s into %s", funcOverview.Name, imagePath))
		}

		data.Rows = append(data.Rows, sideBySideRow{
			Code:  code,
			Image: &imagePath,
		})
	}

	t, err := template.New("side-by-side").Parse(sideBySideTemplate)
	if err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	err = t.Execute(&out, data)
	if err != nil {
		return errors.Wrap(err, "Failed executing side-by-side template")
	}

	indexPath := path.Join(outPath, "index.html")
	err = ioutil.WriteFile(indexPath, out.Bytes(), 0o666)
	if err != nil {
		return errors.Wrap(err, "Failed writing output to "+indexPath)
	}

	return nil
}
