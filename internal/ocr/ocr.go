package ocr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ExtractText runs Tesseract OCR on imageBytes and returns cleaned extracted text.
// Requires the tesseract binary to be installed (brew install tesseract).
func ExtractText(imageBytes []byte) (string, error) {
	tmp, err := os.CreateTemp("", "inti-ocr-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(imageBytes); err != nil {
		tmp.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	outBase := tmp.Name() + "-out"
	defer os.Remove(outBase + ".txt")

	cmd := exec.Command("tesseract", tmp.Name(), outBase)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tesseract: %w\n%s", err, out)
	}

	data, err := os.ReadFile(filepath.Clean(outBase + ".txt"))
	if err != nil {
		return "", fmt.Errorf("read output: %w", err)
	}

	return cleanText(string(data)), nil
}

var multiNewline = regexp.MustCompile(`\n{2,}`)

// cleanText removes OCR word-wrap artifacts and collapses excess blank lines.
//
// Rules:
//   - Within a block (lines separated by a single \n), join lines with a space
//     unless the previous line ends with sentence-closing punctuation.
//   - Blocks separated by 2+ newlines are kept as separate lines (joined with \n).
func cleanText(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	blocks := multiNewline.Split(raw, -1)

	var result []string
	for _, block := range blocks {
		lines := strings.Split(block, "\n")

		var buf strings.Builder
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if buf.Len() == 0 {
				buf.WriteString(line)
				continue
			}
			// Join with space unless the previous content ends a sentence.
			prev := buf.String()
			last := rune(prev[len(prev)-1])
			if last == '.' || last == '!' || last == '?' || last == ':' ||
				last == '"' || last == '\u201d' || last == '\u2019' {
				buf.WriteByte('\n')
			} else {
				buf.WriteByte(' ')
			}
			buf.WriteString(line)
		}

		if t := strings.TrimSpace(buf.String()); t != "" {
			result = append(result, t)
		}
	}

	return strings.Join(result, "\n")
}
