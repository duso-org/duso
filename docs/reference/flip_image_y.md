# flip_image_y

Flips an image vertically (top-bottom mirror across horizontal axis).

## Syntax

```duso
flip_image_y(image)
```

## Parameters

- `image` - A `binary` value containing image data (PNG, JPEG, or GIF)

## Returns

A new `binary` value containing the flipped image with metadata preserved (width, height, format, content_type).

## Description

Mirrors the image vertically, creating a top-bottom flip. The top becomes the bottom and vice versa. Dimensions remain unchanged.

## Format Support

Supports PNG, JPEG, and GIF formats. Output format matches input format.

## Examples

### Basic vertical flip

```duso
image = load_image("photo.jpg")
flipped = flip_image_y(image)
save_image(flipped, "flipped.jpg")
```

### Create upside-down version

```duso
portrait = load_image("portrait.png")
upside_down = flip_image_y(portrait)
save_image(upside_down, "upside_down.png")
```

### Combine with horizontal flip for 180-degree rotation

```duso
image = load_image("photo.jpg")
result = flip_image_y(image) |> flip_image_x()
save_image(result, "rotated_180.jpg")
```

## Behavior

- **Preserves dimensions** - Width and height unchanged
- **Top-bottom mirror** - Top pixels move to bottom, bottom pixels move to top
- **Preserves alpha** - Transparency is preserved

## Metadata

The returned binary includes metadata:

- `width` - Image width in pixels (unchanged)
- `height` - Image height in pixels (unchanged)
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from input if present

## See Also

- [flip_image_x() - Flip horizontally](/docs/reference/flip_image_x.md)
- [rotate_image() - Rotate images](/docs/reference/rotate_image.md)
- [composite_image() - Combine images](/docs/reference/composite_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
