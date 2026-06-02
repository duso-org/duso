# set_image_opacity

Sets the opacity of an image to an absolute value.

## Syntax

```duso
set_image_opacity(image, opacity)
```

## Parameters

- `image` - A `binary` value containing image data (PNG, JPEG, or GIF)
- `opacity` - A number between 0.0 and 1.0 where 0.0 is fully transparent and 1.0 is fully opaque

## Returns

A new `binary` value with the opacity applied to all pixels (width, height, format, content_type preserved).

## Description

Sets the opacity of an entire image to a specific value by multiplying the alpha channel. Use this to make an image uniformly transparent or opaque. For PNG and GIF, the alpha channel is modified; for JPEG, a new alpha channel is created if needed.

## Format Support

Supports PNG, JPEG, and GIF formats. Output format matches input format.

## Examples

### Make image semi-transparent

```duso
image = load_image("photo.png")
transparent = set_image_opacity(image, 0.5)
save_image(transparent, "semi_transparent.png")
```

### Create fully opaque image

```duso
image = load_image("semi_transparent.png")
opaque = set_image_opacity(image, 1.0)
save_image(opaque, "opaque.png")
```

### Make image fully transparent

```duso
image = load_image("photo.png")
invisible = set_image_opacity(image, 0.0)
save_image(invisible, "invisible.png")
```

### Use with composite_image for transparent overlay

```duso
base = load_image("background.jpg")
overlay = load_image("overlay.png")
faded_overlay = set_image_opacity(overlay, 0.3)
result = composite_image(base, faded_overlay, 0, 0)
save_image(result, "composited.jpg")
```

## Behavior

- **Absolute value** - Sets opacity to exact value, not relative to current
- **Range 0.0-1.0** - Values outside this range throw an error
- **Alpha channel creation** - Creates alpha channel for JPEG if needed
- **All pixels affected** - Applied uniformly to entire image

## Metadata

The returned binary includes metadata:

- `width` - Image width in pixels (unchanged)
- `height` - Image height in pixels (unchanged)
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from input if present

## Difference from adjust_image_opacity

- `set_image_opacity()` - Sets absolute opacity (overrides current)
- `adjust_image_opacity()` - Multiplies current opacity by factor (relative)

## See Also

- [adjust_image_opacity() - Adjust opacity relatively](/docs/reference/adjust_image_opacity.md)
- [composite_image() - Combine images](/docs/reference/composite_image.md)
- [grayscale_image() - Convert to grayscale](/docs/reference/grayscale_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
