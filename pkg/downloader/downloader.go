package downloader

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"

	"codeberg.org/go-pdf/fpdf"
)

// DPI used for converting pixels to mm
const dpi = 96.0

// DownloadPdfFromImageSrcs downloads image URLs and creates a PDF with each image as a full-page
func DownloadPdfFromImageSrcs(imgSrcs []string, title string) ([]byte, error) {
	if len(imgSrcs) == 0 {
		return nil, fmt.Errorf("no image sources provided")
	}

	pdf := fpdf.NewCustom(&fpdf.InitType{
		UnitStr: "mm",
		Size:    fpdf.SizeType{}, // dynamic per page
	})
	pdf.SetTitle(title, false)

	for i, src := range imgSrcs {
		// Download image
		resp, err := http.Get(src)
		if err != nil {
			return nil, fmt.Errorf("error fetching image %d: %v", i+1, err)
		}
		imgData, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read image %d: %v", i+1, err)
		}

		// Detect type and dimensions
		imgType := detectImageType(imgData)
		if imgType == "" {
			return nil, fmt.Errorf("unsupported or unknown image type for image %d", i+1)
		}

		cfg, _, err := image.DecodeConfig(bytes.NewReader(imgData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode image %d: %v", i+1, err)
		}

		// Convert pixel size to mm
		widthMM := float64(cfg.Width) * 25.4 / dpi
		heightMM := float64(cfg.Height) * 25.4 / dpi

		// Add page with matching size
		pdf.AddPageFormat("P", fpdf.SizeType{Wd: widthMM, Ht: heightMM})

		// Register image
		alias := fmt.Sprintf("img%d", i)
		options := fpdf.ImageOptions{
			ImageType: imgType,
			ReadDpi:   false,
		}
		pdf.RegisterImageOptionsReader(alias, options, bytes.NewReader(imgData))

		// Add image full page
		pdf.ImageOptions(alias, 0, 0, widthMM, heightMM, false, options, 0, "")
	}

	// Output PDF as bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	fmt.Printf("PDF saved as %s.pdf\n", title)
	return buf.Bytes(), nil
}

// detectImageType detects MIME type using content sniffing
func detectImageType(data []byte) string {
	switch http.DetectContentType(data) {
	case "image/jpeg":
		return "JPG"
	case "image/png":
		return "PNG"
	default:
		return ""
	}
}
