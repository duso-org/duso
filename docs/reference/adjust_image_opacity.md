# adjust_image_opacity

Multiplies the opacity of an image by a factor (relative adjustment).

## Syntax

```duso
adjust_image_opacity(image, opacity)
```

## Parameters

- `image` - A `binary` value containing image data (PNG, JPEG, or GIF)
- `opacity` - A multiplier between 0.0 and 1.0 to apply to current opacity

## Returns

A new `binary` value with adjusted opacity (width, height, format, content_type preserved).

## Description

Adjusts the opacity of an image relative to its current state by multiplying the alpha channel. This is useful for progressively fading images or making semi-transparent images more or less visible. For PNG and GIF, the alpha channel is modified; for JPEG, a new alpha channel is created if needed.

## Format Support

Supports PNG, JPEG, and GIF formats. Output format matches input format.

## Examples

### Make image 50% of its current opacity

```duso
img = load_image("photo.png")
img = adjust_image_opacity(img, 0.5)
save_image(img, "faded.png")
```

### Make already-transparent image more opaque

```duso
img = load_image("semi_transparent.png")
img = adjust_image_opacity(img, 2.0)
save_image(img, "more_visible.png")
```

### Fade to completely transparent

```duso
img = load_image("photo.png")
img = adjust_image_opacity(img, 0.0)
save_image(img, "invisible.png")
```

### Progressive fade effect (multiple images)

```duso
image = load_image("photo.jpg")
fade1 = adjust_image_opacity(image, 0.75)
fade2 = adjust_image_opacity(fade1, 0.75)
fade3 = adjust_image_opacity(fade2, 0.75)
```

### Enhance opacity of translucent overlay

```duso
ovr = load_image("overlay.png")
ovr = adjust_image_opacity(ovr, 1.5)
base = load_image("background.jpg")
base = composite_image(base, ovr, 0, 0)
save_image(base, "composited.jpg")
```

## Behavior

- **Relative adjustment** - Multiplies current opacity by factor
- **Unbounded range** - Values > 1.0 increase opacity, < 1.0 decrease it
- **Clamping** - Results are clamped to valid range [0.0, 1.0]
- **All pixels affected** - Applied uniformly to entire image

## Metadata

The returned binary includes metadata:

- `width` - Image width in pixels (unchanged)
- `height` - Image height in pixels (unchanged)
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from input if present

## Difference from set_image_opacity

- `set_image_opacity()` - Sets absolute opacity (overrides current)
- `adjust_image_opacity()` - Multiplies current opacity by factor (relative)

## Examples of Adjustment Factors

- `0.0` - Result is fully transparent
- `0.5` - Result is half as opaque
- `1.0` - No change to opacity
- `1.5` - Increases opacity by 50% (clamped to max 1.0)
- `2.0` - Doubles opacity (clamped to max 1.0)

## See Also

- [set_image_opacity() - Set absolute opacity](/docs/reference/set_image_opacity.md)
- [composite_image() - Combine images](/docs/reference/composite_image.md)
- [grayscale_image() - Convert to grayscale](/docs/reference/grayscale_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
