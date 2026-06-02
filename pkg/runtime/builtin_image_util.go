package runtime

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// detectImageFormat attempts to detect image format from magic bytes
func detectImageFormat(data []byte) string {
	if len(data) < 4 {
		return "unknown"
	}

	// PNG magic bytes: 89 50 4E 47
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}

	// JPEG magic bytes: FF D8
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return "jpeg"
	}

	// GIF magic bytes: 47 49 46 (GIF)
	if len(data) >= 3 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif"
	}

	return "unknown"
}

// decodeImage decodes binary data into an image.Image, returning format and error
func decodeImage(data []byte) (image.Image, string, error) {
	format := detectImageFormat(data)

	// Try the detected format first
	switch format {
	case "png":
		img, err := png.Decode(bytes.NewReader(data))
		if err == nil {
			return img, "png", nil
		}
	case "jpeg":
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err == nil {
			return img, "jpeg", nil
		}
	case "gif":
		img, err := gif.Decode(bytes.NewReader(data))
		if err == nil {
			return img, "gif", nil
		}
	}

	// Try all formats as fallback
	if img, err := png.Decode(bytes.NewReader(data)); err == nil {
		return img, "png", nil
	}
	if img, err := jpeg.Decode(bytes.NewReader(data)); err == nil {
		return img, "jpeg", nil
	}
	if img, err := gif.Decode(bytes.NewReader(data)); err == nil {
		return img, "gif", nil
	}

	return nil, "", fmt.Errorf("failed to decode image: unsupported format")
}

// encodeImage encodes an image.Image to bytes in the specified format
func encodeImage(img image.Image, format string) ([]byte, error) {
	buf := new(bytes.Buffer)

	switch strings.ToLower(format) {
	case "png":
		err := png.Encode(buf, img)
		return buf.Bytes(), err
	case "jpeg", "jpg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
		return buf.Bytes(), err
	case "gif":
		err := gif.Encode(buf, img, nil)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
}

// setImageMetadata sets standard image metadata on a binary value
func setImageMetadata(bin *script.BinaryValue, img image.Image, format string) {
	bounds := img.Bounds()
	bin.Metadata["width"] = script.NewNumber(float64(bounds.Dx()))
	bin.Metadata["height"] = script.NewNumber(float64(bounds.Dy()))
	bin.Metadata["format"] = script.NewString(format)

	contentType := "image/png"
	if format == "jpeg" {
		contentType = "image/jpeg"
	} else if format == "gif" {
		contentType = "image/gif"
	}
	bin.Metadata["content_type"] = script.NewString(contentType)
}

// extractBinaryImage extracts a binary image from function arguments
func extractBinaryImage(args map[string]any, argKey string) *script.BinaryValue {
	if b, ok := args[argKey]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			return val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			return val.Val.AsBinary()
		}
	}
	return nil
}
