# composite_image

Composites an overlay image on top of a base image with various blend modes.

## Syntax

```duso
composite_image(base, overlay, x, y, blend)
```

## Parameters

- `base` - A `binary` value containing the background image
- `overlay` - A `binary` value containing the image to composite on top
- `x` - Horizontal offset in pixels (default: 0, optional)
- `y` - Vertical offset in pixels (default: 0, optional)
- `blend` - Blend mode as a string (default: "over", optional). Allowed values: `"over"`, `"multiply"`, `"add"`, `"screen"`

## Returns

A new `binary` value with the dimensions of the base image and composited content.

## Description

Layers an overlay image on top of a base image at the specified coordinates. Different blend modes control how the overlay pixels interact with the base. The result always has the dimensions of the base image.

## Format Support

Supports PNG, JPEG, and GIF formats. Both input images must be the same format, and output format matches the base image.

## Blend Modes

- **"over"** (default) - Standard alpha blending. Overlay pixels with alpha combine with base
- **"multiply"** - Multiplies colors: result is darker. Good for shadows
- **"add"** - Adds color values. Good for light effects
- **"screen"** - Inverted multiply. Good for light overlays

## Examples

### Simple overlay at origin

```duso
background = load_image("background.jpg")
logo = load_image("logo.png")
result = composite_image(background, logo)
save_image(result, "with_logo.jpg")
```

### Position overlay at coordinates

```duso
background = load_image("background.jpg")
watermark = load_image("watermark.png")
result = composite_image(background, watermark, 10, 10)
save_image(result, "watermarked.jpg")
```

### Using blend modes

```duso
base = load_image("base.jpg")
shadow = load_image("shadow.png")
with_shadow = composite_image(base, shadow, 5, 5, "multiply")
save_image(with_shadow, "shadowed.jpg")
```

### Light overlay with screen blend

```duso
image = load_image("photo.jpg")
light_effect = load_image("light.png")
result = composite_image(image, light_effect, 0, 0, "screen")
save_image(result, "bright.jpg")
```

### Named parameters

```duso
background = load_image("bg.png")
overlay = load_image("overlay.png")
result = composite_image(background, overlay, x = 50, y = 50, blend = "multiply")
save_image(result, "composite.png")
```

## Behavior

- **Clipping** - Overlay regions beyond base bounds are clipped
- **Dimensions** - Result is always base image size
- **Format matching** - Both inputs should be compatible formats
- **Alpha handling** - Varies by blend mode

## Metadata

The returned binary includes metadata from the base image:

- `width` - Width of base image in pixels
- `height` - Height of base image in pixels
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from base if present

## Performance Notes

- Suitable for watermarking, creating mosaics, and layer composition
- Blend modes may be slower than "over" blending
- Memory-efficient: creates new binary only for result

## See Also

- [flip_image_x() - Flip horizontally](/docs/reference/flip_image_x.md)
- [flip_image_y() - Flip vertically](/docs/reference/flip_image_y.md)
- [crop_image() - Extract regions](/docs/reference/crop_image.md)
- [scale_image() - Resize images](/docs/reference/scale_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
