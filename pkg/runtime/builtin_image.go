package runtime

import (
	"fmt"
	"image"
	"image/draw"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// scaleImage scales src to fill dst using nearest-neighbor sampling
func scaleImage(dst *image.NRGBA, src image.Image) {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	dstBounds := dst.Bounds()
	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()

	// Nearest-neighbor scaling
	for y := 0; y < dstH; y++ {
		srcY := (y * srcH) / dstH
		for x := 0; x < dstW; x++ {
			srcX := (x * srcW) / dstW
			dst.Set(x, y, src.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
		}
	}
}

// builtinScaleImage scales an image to fit/fill/stretch to given dimensions
func builtinScaleImage(evaluator *Evaluator, args map[string]any) (any, error) {
	// Extract binary image
	var binary *script.BinaryValue
	if b, ok := args["0"]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			binary = val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			binary = val.Val.AsBinary()
		}
	} else if b, ok := args["image"]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			binary = val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			binary = val.Val.AsBinary()
		}
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("scale_image() requires a binary image as first argument")
	}

	// Extract max_x
	var maxX float64
	if x, ok := args["1"].(float64); ok {
		maxX = x
	} else if x, ok := args["max_x"].(float64); ok {
		maxX = x
	} else {
		return nil, fmt.Errorf("scale_image() requires max_x (width) as number")
	}

	// Extract max_y
	var maxY float64
	if y, ok := args["2"].(float64); ok {
		maxY = y
	} else if y, ok := args["max_y"].(float64); ok {
		maxY = y
	} else {
		return nil, fmt.Errorf("scale_image() requires max_y (height) as number")
	}

	if maxX <= 0 || maxY <= 0 {
		return nil, fmt.Errorf("scale_image() dimensions must be positive")
	}

	// Extract mode (default "fit")
	mode := "fit"
	if m, ok := args["3"].(string); ok {
		mode = m
	} else if m, ok := args["mode"].(string); ok {
		mode = m
	}
	mode = strings.ToLower(mode)

	if mode != "fit" && mode != "fill" && mode != "stretch" {
		return nil, fmt.Errorf("scale_image() invalid mode '%s': use 'fit', 'fill', or 'stretch'", mode)
	}

	// Decode image
	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("scale_image() %w", err)
	}

	bounds := img.Bounds()
	origWidth := float64(bounds.Dx())
	origHeight := float64(bounds.Dy())

	var newWidth, newHeight int

	switch mode {
	case "stretch":
		// Force to exact dimensions
		newWidth = int(maxX)
		newHeight = int(maxY)

	case "fit":
		// Scale to fit within bounds, preserve aspect ratio
		scaleX := maxX / origWidth
		scaleY := maxY / origHeight
		scale := scaleX
		if scaleY < scale {
			scale = scaleY
		}
		newWidth = int(origWidth * scale)
		newHeight = int(origHeight * scale)

	case "fill":
		// Scale to fill bounds, preserve aspect ratio (may crop)
		scaleX := maxX / origWidth
		scaleY := maxY / origHeight
		scale := scaleX
		if scaleY > scale {
			scale = scaleY
		}
		newWidth = int(origWidth * scale)
		newHeight = int(origHeight * scale)
	}

	// Create new image with target dimensions
	dst := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Scale using nearest-neighbor sampling
	scaleImage(dst, img)

	// For "fill" mode, crop to exact dimensions from center
	if mode == "fill" && (newWidth > int(maxX) || newHeight > int(maxY)) {
		offsetX := (newWidth - int(maxX)) / 2
		offsetY := (newHeight - int(maxY)) / 2
		cropBounds := image.Rect(offsetX, offsetY, offsetX+int(maxX), offsetY+int(maxY))
		cropped := dst.SubImage(cropBounds).(*image.NRGBA)
		dst = cropped
	}

	// Encode result
	encoded, err := encodeImage(dst, format)
	if err != nil {
		return nil, fmt.Errorf("scale_image() encode failed: %w", err)
	}

	// Create result binary with metadata
	result := script.NewBinary(encoded)
	resultBin := result.AsBinary()
	resultBin.Metadata["width"] = script.NewNumber(float64(dst.Bounds().Dx()))
	resultBin.Metadata["height"] = script.NewNumber(float64(dst.Bounds().Dy()))
	resultBin.Metadata["format"] = script.NewString(format)

	contentType := "image/png"
	if format == "jpeg" {
		contentType = "image/jpeg"
	} else if format == "gif" {
		contentType = "image/gif"
	}
	resultBin.Metadata["content_type"] = script.NewString(contentType)

	return result, nil
}

// builtinCropImage crops an image to a specified region
func builtinCropImage(evaluator *Evaluator, args map[string]any) (any, error) {
	// Extract binary image
	var binary *script.BinaryValue
	if b, ok := args["0"]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			binary = val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			binary = val.Val.AsBinary()
		}
	} else if b, ok := args["image"]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			binary = val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			binary = val.Val.AsBinary()
		}
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("crop_image() requires a binary image as first argument")
	}

	// Extract crop coordinates and dimensions
	var x, y, width, height float64

	if v, ok := args["1"].(float64); ok {
		x = v
	} else if v, ok := args["x"].(float64); ok {
		x = v
	} else {
		return nil, fmt.Errorf("crop_image() requires x coordinate")
	}

	if v, ok := args["2"].(float64); ok {
		y = v
	} else if v, ok := args["y"].(float64); ok {
		y = v
	} else {
		return nil, fmt.Errorf("crop_image() requires y coordinate")
	}

	if v, ok := args["3"].(float64); ok {
		width = v
	} else if v, ok := args["width"].(float64); ok {
		width = v
	} else {
		return nil, fmt.Errorf("crop_image() requires crop width")
	}

	if v, ok := args["4"].(float64); ok {
		height = v
	} else if v, ok := args["height"].(float64); ok {
		height = v
	} else {
		return nil, fmt.Errorf("crop_image() requires crop height")
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("crop_image() width and height must be positive")
	}

	// Decode image
	img, format, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("crop_image() %w", err)
	}

	// Create crop rectangle
	cropBounds := image.Rect(int(x), int(y), int(x+width), int(y+height))

	// Clip to image bounds
	imgBounds := img.Bounds()
	cropBounds = cropBounds.Intersect(imgBounds)

	if cropBounds.Empty() {
		return nil, fmt.Errorf("crop_image() crop region outside image bounds")
	}

	// Extract cropped region
	cropped := img.(interface{ SubImage(r image.Rectangle) image.Image }).SubImage(cropBounds)

	// For NRGBA conversion if needed
	var result image.Image = cropped
	if _, ok := cropped.(*image.NRGBA); !ok {
		// Convert to NRGBA for consistent handling
		dst := image.NewNRGBA(image.Rect(0, 0, cropBounds.Dx(), cropBounds.Dy()))
		draw.Draw(dst, dst.Bounds(), cropped, cropBounds.Min, draw.Src)
		result = dst
	}

	// Encode result
	encoded, err := encodeImage(result, format)
	if err != nil {
		return nil, fmt.Errorf("crop_image() encode failed: %w", err)
	}

	// Create result binary with metadata
	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	bin.Metadata["width"] = script.NewNumber(float64(cropBounds.Dx()))
	bin.Metadata["height"] = script.NewNumber(float64(cropBounds.Dy()))
	bin.Metadata["format"] = script.NewString(format)

	contentType := "image/png"
	if format == "jpeg" {
		contentType = "image/jpeg"
	} else if format == "gif" {
		contentType = "image/gif"
	}
	bin.Metadata["content_type"] = script.NewString(contentType)

	return resultBin, nil
}

// builtinConvertImage converts an image to a different format
func builtinConvertImage(evaluator *Evaluator, args map[string]any) (any, error) {
	// Extract binary image
	var binary *script.BinaryValue
	if b, ok := args["0"]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			binary = val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			binary = val.Val.AsBinary()
		}
	} else if b, ok := args["image"]; ok {
		if val, ok := b.(script.Value); ok && val.IsBinary() {
			binary = val.AsBinary()
		} else if val, ok := b.(*script.ValueRef); ok && val.Val.IsBinary() {
			binary = val.Val.AsBinary()
		}
	}

	if binary == nil || binary.Data == nil {
		return nil, fmt.Errorf("convert_image() requires a binary image as first argument")
	}

	// Extract target format
	var format string
	if f, ok := args["1"].(string); ok {
		format = f
	} else if f, ok := args["format"].(string); ok {
		format = f
	} else {
		return nil, fmt.Errorf("convert_image() requires target format (png, jpeg, gif)")
	}

	format = strings.ToLower(format)
	if format == "jpg" {
		format = "jpeg"
	}

	if format != "png" && format != "jpeg" && format != "gif" {
		return nil, fmt.Errorf("convert_image() unsupported format '%s': use 'png', 'jpeg', or 'gif'", format)
	}

	// Decode image
	img, _, err := decodeImage(*binary.Data)
	if err != nil {
		return nil, fmt.Errorf("convert_image() %w", err)
	}

	// For JPEG encoding, convert to NRGBA if needed for consistency
	if format == "jpeg" {
		bounds := img.Bounds()
		dst := image.NewNRGBA(bounds)
		draw.Draw(dst, bounds, img, bounds.Min, draw.Src)
		img = dst
	}

	// Encode to target format
	encoded, err := encodeImage(img, format)
	if err != nil {
		return nil, fmt.Errorf("convert_image() encode failed: %w", err)
	}

	// Create result binary with metadata
	result := script.NewBinary(encoded)
	resultBin := result.AsBinary()

	// Preserve dimensions
	bounds := img.Bounds()
	resultBin.Metadata["width"] = script.NewNumber(float64(bounds.Dx()))
	resultBin.Metadata["height"] = script.NewNumber(float64(bounds.Dy()))
	resultBin.Metadata["format"] = script.NewString(format)

	contentType := "image/png"
	if format == "jpeg" {
		contentType = "image/jpeg"
	} else if format == "gif" {
		contentType = "image/gif"
	}
	resultBin.Metadata["content_type"] = script.NewString(contentType)

	return result, nil
}
