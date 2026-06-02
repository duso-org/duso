package runtime

import (
	"fmt"
	"image"
	"image/color"

	"github.com/duso-org/duso/pkg/script"
)

// builtinGrayscaleImage converts an image to grayscale
func builtinGrayscaleImage(evaluator *Evaluator, args map[string]any) (any, error) {
	binary := extractBinaryImage(args, "0")
	if binary == nil {
		binary = extractBinaryImage(args, "image")
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("grayscale_image() requires a binary image as argument")
	}

	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("grayscale_image() %w", err)
	}

	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// Un-premultiply RGB values (RGBA returns 16-bit pre-multiplied)
			var r8, g8, b8 uint8
			if a == 0 {
				r8, g8, b8 = 0, 0, 0
			} else {
				r8 = uint8((r * 255) / a)
				g8 = uint8((g * 255) / a)
				b8 = uint8((b * 255) / a)
			}

			// Luminance formula
			gray := uint8(float64(r8)*0.299 + float64(g8)*0.587 + float64(b8)*0.114)

			// Preserve original alpha directly
			dst.Set(x, y, color.RGBA{gray, gray, gray, uint8(a >> 8)})
		}
	}

	encoded, err := encodeImage(dst, format)
	if err != nil {
		return nil, fmt.Errorf("grayscale_image() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, dst, format)

	return resultBin, nil
}

// builtinSetImageOpacity sets the opacity of an image to an absolute value (0-1)
// set_image_opacity(image, opacity) or set_image_opacity(image, opacity=0.5)
func builtinSetImageOpacity(evaluator *Evaluator, args map[string]any) (any, error) {
	binary := extractBinaryImage(args, "0")
	if binary == nil {
		binary = extractBinaryImage(args, "image")
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("set_image_opacity() requires a binary image as first argument")
	}

	var opacity float64
	if o, ok := args["opacity"].(float64); ok {
		opacity = o
	} else if o, ok := args["1"].(float64); ok {
		opacity = o
	} else {
		return nil, fmt.Errorf("set_image_opacity() requires opacity (0.0-1.0)")
	}

	if opacity < 0 || opacity > 1 {
		return nil, fmt.Errorf("set_image_opacity() opacity must be between 0 and 1")
	}

	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("set_image_opacity() %w", err)
	}

	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)

	opacityAlpha := uint8(opacity * 255)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			dst.Set(x, y, color.NRGBA{
				uint8(r >> 8),
				uint8(g >> 8),
				uint8(b >> 8),
				opacityAlpha,
			})
		}
	}

	encoded, err := encodeImage(dst, format)
	if err != nil {
		return nil, fmt.Errorf("set_image_opacity() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, dst, format)

	return resultBin, nil
}

// builtinAdjustImageOpacity multiplies the opacity of an image (0-1)
// adjust_image_opacity(image, opacity) or adjust_image_opacity(image, opacity=0.5)
func builtinAdjustImageOpacity(evaluator *Evaluator, args map[string]any) (any, error) {
	binary := extractBinaryImage(args, "0")
	if binary == nil {
		binary = extractBinaryImage(args, "image")
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("adjust_image_opacity() requires a binary image as first argument")
	}

	var opacity float64
	if o, ok := args["opacity"].(float64); ok {
		opacity = o
	} else if o, ok := args["1"].(float64); ok {
		opacity = o
	} else {
		return nil, fmt.Errorf("adjust_image_opacity() requires opacity (0.0-1.0)")
	}

	if opacity < 0 || opacity > 1 {
		return nil, fmt.Errorf("adjust_image_opacity() opacity must be between 0 and 1")
	}

	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("adjust_image_opacity() %w", err)
	}

	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			newA := uint8((float64(a>>8) * opacity))
			dst.Set(x, y, color.NRGBA{
				uint8(r >> 8),
				uint8(g >> 8),
				uint8(b >> 8),
				newA,
			})
		}
	}

	encoded, err := encodeImage(dst, format)
	if err != nil {
		return nil, fmt.Errorf("adjust_image_opacity() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, dst, format)

	return resultBin, nil
}
