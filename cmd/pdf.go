package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/100nandoo/inti/internal/pdf"
	"github.com/spf13/cobra"
)

var pdfCmd = &cobra.Command{
	Use:   "pdf <input.pdf>",
	Short: "Convert PDF pages to PNG images",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]

		if strings.ToLower(filepath.Ext(inputPath)) != ".pdf" {
			return fmt.Errorf("input file must have a .pdf extension")
		}

		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			base := filepath.Base(inputPath)
			outputDir = strings.TrimSuffix(base, filepath.Ext(base))
		}

		n, err := pdf.Convert(inputPath, outputDir)
		if err != nil {
			return err
		}

		fmt.Printf("converted %d pages → %s/\n", n, outputDir)
		return nil
	},
}

func init() {
	pdfCmd.Flags().String("output", "", "output directory (default: PDF filename without extension)")
}
