package docs

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/NetSys/quilt/stitch"
	"github.com/NetSys/quilt/util"
)

const (
	blockStart     = "```javascript\n"
	blockEnd       = "```\n"
	commentPattern = "^\\[//\\]: # \\((.*)\\)\\W*$"
)

var errUnbalanced = errors.New("unbalanced code blocks")

type readmeParser struct {
	currentBlock string
	// Map block ID to code block.
	codeBlocks map[string]string
	recording  bool
}

func (parser *readmeParser) parse(line string) error {
	isStart := line == blockStart
	isEnd := line == blockEnd
	reComment := regexp.MustCompile(commentPattern)
	match := reComment.FindStringSubmatch(line)
	isComment := len(match) > 0

	if (isStart && parser.recording) || (isEnd && !parser.recording) {
		return errUnbalanced
	}

	switch {
	case isComment:
		parser.currentBlock = match[1]
	case isStart:
		parser.recording = true

		if parser.currentBlock == "" {
			return errors.New("missing code block id")
		}

		if _, ok := parser.codeBlocks[parser.currentBlock]; !ok {
			parser.codeBlocks[parser.currentBlock] = ""
		}
	case isEnd:
		parser.recording = false
		parser.currentBlock = ""
	}

	if parser.recording && !isStart {
		parser.codeBlocks[parser.currentBlock] += line
	}

	return nil
}

func (parser readmeParser) blocks() (map[string]string, error) {
	if parser.recording {
		return nil, errUnbalanced
	}
	return parser.codeBlocks, nil
}

func TestReadme(t *testing.T) {
	f, err := util.Open("../README.md")
	if err != nil {
		t.Errorf("Failed to open README: %s", err.Error())
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	parser := readmeParser{}
	parser.codeBlocks = make(map[string]string)

	for scanner.Scan() {
		if err := parser.parse(scanner.Text() + "\n"); err != nil {
			t.Errorf("Failed to parse README: %s", err.Error())
			return
		}
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Failed to read README: %s", err.Error())
		return
	}

	blocks, err := parser.blocks()
	if err != nil {
		t.Errorf("Failed to parse README: %s", err.Error())
		return
	}

	goPath := os.Getenv("GOPATH")
	quiltPath := filepath.Join(goPath, "src")

	for _, block := range blocks {
		if err = checkConfig(block, quiltPath); err != nil {
			t.Errorf(err.Error())
		}
	}
}

func checkConfig(content string, quiltPath string) error {
	oldHTTPGet := stitch.HTTPGet
	defer func() {
		stitch.HTTPGet = oldHTTPGet
	}()

	stitch.HTTPGet = func(url string) (*http.Response, error) {
		resp := http.Response{
			Body: ioutil.NopCloser(bytes.NewBufferString("")),
		}
		return &resp, nil
	}
	_, err := stitch.FromJavascript(content, stitch.ImportGetter{
		Path: quiltPath,
	})
	if err != nil {
		return err
	}
	return nil
}
