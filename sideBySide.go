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
	"path/filepath"
	"sort"
	"strings"
)

//go:embed side-by-side.html.tmpl
var sideBySideTemplate string

//go:embed index.html.tmpl
var sideBySideIndexTemplate string

var templates struct {
	File  *template.Template
	Index *template.Template
}

func init() {
	t, err := template.New("side-by-side").Parse(sideBySideTemplate)
	if err != nil {
		log.Fatal(err)
	}
	templates.File = t

	t, err = template.New("index").Parse(sideBySideIndexTemplate)
	if err != nil {
		log.Fatal(err)
	}
	templates.Index = t
}

type sideBySideRow struct {
	Name  string
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

	baseName := filepath.Base(fileOverview.Filename)

	data := sideBySideData{
		Filename: baseName,
		Rows:     nil,
	}
	var lineBlocks []string
	currentLine := 0

	for _, overview := range functionOverviews {
		lineBlocks = append(lineBlocks, strings.Join(lines[currentLine:overview.Line-1], "\n"))
		currentLine = overview.Line - 1
	}
	lineBlocks = append(lineBlocks, strings.Join(lines[currentLine:], "\n"))

	data.Rows = append(data.Rows, sideBySideRow{
		Code:  lineBlocks[0],
		Image: nil,
	})

	log.Printf("%s -> %s", fileOverview.Filename, baseName)

	for i, code := range lineBlocks[1:] {
		funcOverview := functionOverviews[i]

		imageName := fmt.Sprintf("%s._.%s.png", baseName, funcOverview.Name)
		imagePath := path.Join(outPath, imageName)
		err := renderPng([]byte(funcOverview.Dot), imagePath)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed rendering PNG for %s into %s", funcOverview.Name, imagePath))
		}

		data.Rows = append(data.Rows, sideBySideRow{
			Name:  funcOverview.Name,
			Code:  code,
			Image: &imageName,
		})
	}

	var out bytes.Buffer
	err = templates.File.Execute(&out, data)
	if err != nil {
		return errors.Wrap(err, "Failed executing side-by-side template")
	}

	indexPath := path.Join(outPath, fmt.Sprintf("%s.html", baseName))
	err = ioutil.WriteFile(indexPath, out.Bytes(), 0o666)
	if err != nil {
		return errors.Wrap(err, "Failed writing output to "+indexPath)
	}

	return nil
}
