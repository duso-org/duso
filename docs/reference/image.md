# Image Functions

Duso provides a comprehensive set of built-in image manipulation functions. All functions work with PNG, JPEG, and GIF formats.

## Loading & Saving

- [load_image()](/docs/reference/load_image.md) - Load image from file
- [save_image()](/docs/reference/save_image.md) - Save image to file

## Resizing & Cropping

- [scale_image()](/docs/reference/scale_image.md) - Resize with fit/fill/stretch modes
- [crop_image()](/docs/reference/crop_image.md) - Extract rectangular region

## Transformations

- [rotate_image()](/docs/reference/rotate_image.md) - Rotate by 90/180/270 degrees
- [flip_image_x()](/docs/reference/flip_image_x.md) - Flip horizontally (left-right mirror)
- [flip_image_y()](/docs/reference/flip_image_y.md) - Flip vertically (top-bottom mirror)

## Effects & Adjustments

- [grayscale_image()](/docs/reference/grayscale_image.md) - Convert to grayscale
- [set_image_opacity()](/docs/reference/set_image_opacity.md) - Set absolute opacity (0.0-1.0)
- [adjust_image_opacity()](/docs/reference/adjust_image_opacity.md) - Multiply opacity by factor

## Composition & Format

- [composite_image()](/docs/reference/composite_image.md) - Layer overlay on base with blend modes
- [convert_image()](/docs/reference/convert_image.md) - Convert between PNG/JPEG/GIF formats

## Data Type

- [binary](/docs/reference/binary.md) - Binary data type for image content and other binary data

## Quick Example

```duso
// Load, transform, and save an image
image = load_image("photo.jpg")
rotated = rotate_image(image, 90)
scaled = scale_image(rotated, 300, 300, "fit")
grayscale = grayscale_image(scaled)
save_image(grayscale, "processed.jpg")

// Create a watermarked image
base = load_image("background.jpg")
watermark = load_image("watermark.png")
faded = set_image_opacity(watermark, 0.3)
result = composite_image(base, faded, 10, 10)
save_image(result, "watermarked.jpg")
```
