package runtime

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/duso-org/duso/pkg/script"
)

// builtinRotateImage rotates an image by 90, 180, or 270 degrees
// rotate_image(image, degrees) or rotate_image(image, degrees=90)
func builtinRotateImage(evaluator *Evaluator, args map[string]any) (any, error) {
	binary := extractBinaryImage(args, "0")
	if binary == nil {
		binary = extractBinaryImage(args, "image")
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("rotate_image() requires a binary image as first argument")
	}

	var degrees float64
	if d, ok := args["degrees"].(float64); ok {
		degrees = d
	} else if d, ok := args["1"].(float64); ok {
		degrees = d
	} else {
		return nil, fmt.Errorf("rotate_image() requires degrees (90, 180, or 270)")
	}

	if degrees != 90 && degrees != 180 && degrees != 270 {
		return nil, fmt.Errorf("rotate_image() degrees must be 90, 180, or 270")
	}

	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("rotate_image() %w", err)
	}

	var rotated image.Image
	steps := int(degrees / 90)

	for i := 0; i < steps; i++ {
		if i == 0 {
			rotated = rotateOnce(img)
		} else {
			rotated = rotateOnce(rotated)
		}
	}

	// Convert to RGBA if needed
	var result image.Image = rotated
	if _, ok := rotated.(*image.RGBA); !ok {
		rbounds := rotated.Bounds()
		rgba := image.NewRGBA(rbounds)
		draw.Draw(rgba, rbounds, rotated, rbounds.Min, draw.Src)
		result = rgba
	}

	encoded, err := encodeImage(result, format)
	if err != nil {
		return nil, fmt.Errorf("rotate_image() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, result, format)

	return resultBin, nil
}

// rotateOnce rotates an image 90 degrees clockwise
func rotateOnce(src image.Image) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, height, width))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newX := height - 1 - (y - bounds.Min.Y)
			newY := x - bounds.Min.X
			dst.Set(newX, newY, src.At(x, y))
		}
	}

	return dst
}

// builtinFlipImageX flips an image horizontally (left-right mirror across vertical axis)
func builtinFlipImageX(evaluator *Evaluator, args map[string]any) (any, error) {
	binary := extractBinaryImage(args, "0")
	if binary == nil {
		binary = extractBinaryImage(args, "image")
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("flip_image_x() requires a binary image as argument")
	}

	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("flip_image_x() %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newX := bounds.Max.X - 1 - (x - bounds.Min.X)
			dst.Set(newX, y-bounds.Min.Y, img.At(x, y))
		}
	}

	encoded, err := encodeImage(dst, format)
	if err != nil {
		return nil, fmt.Errorf("flip_image_x() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, dst, format)

	return resultBin, nil
}

// builtinFlipImageY flips an image vertically (top-bottom mirror across horizontal axis)
func builtinFlipImageY(evaluator *Evaluator, args map[string]any) (any, error) {
	binary := extractBinaryImage(args, "0")
	if binary == nil {
		binary = extractBinaryImage(args, "image")
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("flip_image_y() requires a binary image as argument")
	}

	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("flip_image_y() %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newY := bounds.Max.Y - 1 - (y - bounds.Min.Y)
			dst.Set(x-bounds.Min.X, newY, img.At(x, y))
		}
	}

	encoded, err := encodeImage(dst, format)
	if err != nil {
		return nil, fmt.Errorf("flip_image_y() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, dst, format)

	return resultBin, nil
}
