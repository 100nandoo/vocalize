package ocr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtractText runs Tesseract OCR on imageBytes and returns the extracted text.
// Requires the tesseract binary to be installed (brew install tesseract).
func ExtractText(imageBytes []byte) (string, error) {
	tmp, err := os.CreateTemp("", "vocalize-ocr-*")
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

	return strings.TrimSpace(string(data)), nil
}
