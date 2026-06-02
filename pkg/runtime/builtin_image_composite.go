package runtime

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// builtinCompositeImage composites an overlay image on top of a base image
func builtinCompositeImage(evaluator *Evaluator, args map[string]any) (any, error) {
	baseBinary := extractBinaryImage(args, "0")
	if baseBinary == nil {
		baseBinary = extractBinaryImage(args, "base")
	}

	if baseBinary == nil || baseBinary.Data == nil {
		return nil, fmt.Errorf("composite_image() requires a base image as first argument")
	}

	overlayBinary := extractBinaryImage(args, "1")
	if overlayBinary == nil {
		overlayBinary = extractBinaryImage(args, "overlay")
	}

	if overlayBinary == nil || overlayBinary.Data == nil {
		return nil, fmt.Errorf("composite_image() requires an overlay image as second argument")
	}

	baseImg, format, err := decodeImage(*baseBinary.Data)
	if err != nil {
		return nil, fmt.Errorf("composite_image() base: %w", err)
	}

	overlayImg, _, err := decodeImage(*overlayBinary.Data)
	if err != nil {
		return nil, fmt.Errorf("composite_image() overlay: %w", err)
	}

	// Extract x, y (named x, y or positional args 2, 3; default 0, 0)
	var x, y float64
	if v, ok := args["x"].(float64); ok {
		x = v
	} else if v, ok := args["2"].(float64); ok {
		x = v
	}

	if v, ok := args["y"].(float64); ok {
		y = v
	} else if v, ok := args["3"].(float64); ok {
		y = v
	}

	// Extract blend mode (named blend or positional arg 4; default "over")
	blendMode := "over"
	if m, ok := args["blend"].(string); ok {
		blendMode = strings.ToLower(m)
	} else if m, ok := args["4"].(string); ok {
		blendMode = strings.ToLower(m)
	}

	if blendMode != "over" && blendMode != "multiply" && blendMode != "add" && blendMode != "screen" {
		return nil, fmt.Errorf("composite_image() invalid blend mode '%s': use 'over', 'multiply', 'add', or 'screen'", blendMode)
	}

	// Create result image as copy of base
	baseBounds := baseImg.Bounds()
	result := image.NewNRGBA(baseBounds)
	draw.Draw(result, baseBounds, baseImg, baseBounds.Min, draw.Src)

	// Composite overlay onto result
	overlayBounds := overlayImg.Bounds()
	offsetX := int(x)
	offsetY := int(y)

	for oy := overlayBounds.Min.Y; oy < overlayBounds.Max.Y; oy++ {
		for ox := overlayBounds.Min.X; ox < overlayBounds.Max.X; ox++ {
			dx := offsetX + (ox - overlayBounds.Min.X)
			dy := offsetY + (oy - overlayBounds.Min.Y)

			if dx >= baseBounds.Min.X && dx < baseBounds.Max.X &&
				dy >= baseBounds.Min.Y && dy < baseBounds.Max.Y {
				baseColor := result.At(dx, dy)
				overlayColor := overlayImg.At(ox, oy)

				// Convert to NRGBA (non-premultiplied)
				basergba := color.NRGBAModel.Convert(baseColor).(color.NRGBA)
				overlayrgba := color.NRGBAModel.Convert(overlayColor).(color.NRGBA)

				// Normalize colors to [0, 1]
				br := float64(basergba.R) / 255.0
				bg := float64(basergba.G) / 255.0
				bb := float64(basergba.B) / 255.0
				ba := float64(basergba.A) / 255.0

				or := float64(overlayrgba.R) / 255.0
				og := float64(overlayrgba.G) / 255.0
				ob := float64(overlayrgba.B) / 255.0
				oa := float64(overlayrgba.A) / 255.0

				var outR, outG, outB, outA float64

				// Non-premultiplied alpha compositing formula:
				// αo = αs + (1 - αs) × αb
				// Co = ((1 - αb) × Cs + αb × B(Cb, Cs)) / αo
				outA = oa + (1-oa)*ba

				if outA > 0 {
					switch blendMode {
					case "over":
						// Standard alpha over: out = src + dst * (1 - src.a)
						outR = (or*oa + br*ba*(1-oa)) / outA
						outG = (og*oa + bg*ba*(1-oa)) / outA
						outB = (ob*oa + bb*ba*(1-oa)) / outA

					case "multiply":
						// Multiply accounting for both base and overlay alpha
						// base*(1-oa)*ba + (base*overlay)*oa*ba + overlay*oa*(1-ba)
						mulR := br * or
						mulG := bg * og
						mulB := bb * ob
						outR = br*(1-oa)*ba + mulR*oa*ba + or*oa*(1-ba)
						outG = bg*(1-oa)*ba + mulG*oa*ba + og*oa*(1-ba)
						outB = bb*(1-oa)*ba + mulB*oa*ba + ob*oa*(1-ba)

					case "add":
						// Add accounting for both base and overlay alpha
						// base*(1-oa)*ba + add*(oa*ba) + overlay*oa*(1-ba)
						addR := br + or
						addG := bg + og
						addB := bb + ob
						if addR > 1 {
							addR = 1
						}
						if addG > 1 {
							addG = 1
						}
						if addB > 1 {
							addB = 1
						}
						outR = br*(1-oa)*ba + addR*oa*ba + or*oa*(1-ba)
						outG = bg*(1-oa)*ba + addG*oa*ba + og*oa*(1-ba)
						outB = bb*(1-oa)*ba + addB*oa*ba + ob*oa*(1-ba)

					case "screen":
						// Screen accounting for both base and overlay alpha
						// base*(1-oa)*ba + screen*(oa*ba) + overlay*oa*(1-ba)
						scrR := 1 - (1-br)*(1-or)
						scrG := 1 - (1-bg)*(1-og)
						scrB := 1 - (1-bb)*(1-ob)
						outR = br*(1-oa)*ba + scrR*oa*ba + or*oa*(1-ba)
						outG = bg*(1-oa)*ba + scrG*oa*ba + og*oa*(1-ba)
						outB = bb*(1-oa)*ba + scrB*oa*ba + ob*oa*(1-ba)
					}
				}

				result.Set(dx, dy, color.NRGBA{
					uint8(outR * 255),
					uint8(outG * 255),
					uint8(outB * 255),
					uint8(outA * 255),
				})
			}
		}
	}

	encoded, err := encodeImage(result, format)
	if err != nil {
		return nil, fmt.Errorf("composite_image() encode failed: %w", err)
	}

	resultBin := script.NewBinary(encoded)
	bin := resultBin.AsBinary()
	setImageMetadata(bin, result, format)

	return resultBin, nil
}
