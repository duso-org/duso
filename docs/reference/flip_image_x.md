# flip_image_x

Flips an image horizontally (left-right mirror across vertical axis).

## Syntax

```duso
flip_image_x(image)
```

## Parameters

- `image` - A `binary` value containing image data (PNG, JPEG, or GIF)

## Returns

A new `binary` value containing the flipped image with metadata preserved (width, height, format, content_type).

## Description

Mirrors the image horizontally, creating a left-right flip. The left side becomes the right side and vice versa. Dimensions remain unchanged.

## Format Support

Supports PNG, JPEG, and GIF formats. Output format matches input format.

## Examples

### Basic horizontal flip

```duso
image = load_image("photo.jpg")
flipped = flip_image_x(image)
save_image(flipped, "flipped.jpg")
```

### Create mirror effect

```duso
portrait = load_image("portrait.png")
mirror = flip_image_x(portrait)
side_by_side = composite_image(portrait, mirror, portrait.width, 0)
save_image(side_by_side, "mirror.png")
```

### Chaining with other operations

```duso
image = load_image("photo.jpg")
flipped = flip_image_x(image)
scaled = scale_image(flipped, 300, 300, "fit")
save_image(scaled, "thumbnail.jpg")
```

## Behavior

- **Preserves dimensions** - Width and height unchanged
- **Left-right mirror** - Left pixels move to right, right pixels move to left
- **Preserves alpha** - Transparency is preserved

## Metadata

The returned binary includes metadata:

- `width` - Image width in pixels (unchanged)
- `height` - Image height in pixels (unchanged)
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from input if present

## See Also

- [flip_image_y() - Flip vertically](/docs/reference/flip_image_y.md)
- [rotate_image() - Rotate images](/docs/reference/rotate_image.md)
- [composite_image() - Combine images](/docs/reference/composite_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
