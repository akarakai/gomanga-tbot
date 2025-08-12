package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)


func DownloadPdfFromImageSrcs(imgSrcs []string, title string) (string, error) {
	if len(imgSrcs) == 0 {
		return "", fmt.Errorf("no image sources provided")
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(title, false)

	for i, src := range imgSrcs {
		resp, err := http.Get(src)
		if err != nil {
			return "", fmt.Errorf("error fetching image %d: %v", i+1, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("image %d: HTTP %d", i+1, resp.StatusCode)
		}

		imgData, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read image %d: %v", i+1, err)
		}

		// Detect image type from URL or headers
		imgType := detectImageType(src, imgData)

		if imgType == "" {
			return "", fmt.Errorf("unsupported or unknown image type for image %d", i+1)
		}

		alias := fmt.Sprintf("img%d", i)
		options := fpdf.ImageOptions{
			ImageType: imgType,
			ReadDpi:   true,
		}

		pdf.RegisterImageOptionsReader(alias, options, bytes.NewReader(imgData))

		pdf.AddPage()
		pdf.ImageOptions(alias, 10, 10, 190, 0, false, options, 0, "")
	}

	// Output PDF
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %v", err)
	}

	outputFile := fmt.Sprintf("%s.pdf", sanitizeFileName(title))
	err = os.WriteFile(outputFile, buf.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write PDF file: %v", err)
	}

	return outputFile, nil
}



func detectImageType(src string, data []byte) string {
	// First try based on extension
	ext := strings.ToLower(filepath.Ext(src))
	switch ext {
	case ".jpg", ".jpeg":
		return "JPG"
	case ".png":
		return "PNG"
	}

	// Fallback to sniffing content
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg":
		return "JPG"
	case "image/png":
		return "PNG"
	default:
		return ""
	}
}


func sanitizeFileName(name string) string {
	// Strip problematic characters for file systems
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return -1
		}
		return r
	}, name)
}