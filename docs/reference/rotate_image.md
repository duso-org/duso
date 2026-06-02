# rotate_image

Rotates an image by 90, 180, or 270 degrees.

## Syntax

```duso
rotate_image(image, degrees)
```

## Parameters

- `image` - A `binary` value containing image data (PNG, JPEG, or GIF)
- `degrees` - Rotation angle: `90`, `180`, or `270`

## Returns

A new `binary` value containing the rotated image with updated metadata (width, height, format, content_type).

## Description

Rotates an image by the specified number of degrees. Only 90-degree increments are supported. For 90 and 270-degree rotations, width and height are swapped.

## Format Support

Supports PNG, JPEG, and GIF formats. Output format matches input format.

## Examples

### Rotate 90 degrees clockwise

```duso
image = load_image("photo.jpg")
rotated = rotate_image(image, 90)
save_image(rotated, "rotated.jpg")
```

### Flip upside down (180 degrees)

```duso
portrait = load_image("portrait.png")
flipped = rotate_image(portrait, 180)
save_image(flipped, "upside_down.png")
```

### Rotate 270 degrees (90 counter-clockwise)

```duso
landscape = load_image("landscape.gif")
rotated = rotate_image(landscape, 270)
save_image(rotated, "rotated.gif")
```

### Chaining operations

```duso
image = load_image("photo.jpg")
result = rotate_image(image, 90)
  |> crop_image(0, 0, 300, 300)
  |> save_image("processed.jpg")
```

## Behavior

- **90 degrees** - Clockwise rotation; dimensions swapped
- **180 degrees** - Upside down; dimensions unchanged
- **270 degrees** - Counter-clockwise rotation; dimensions swapped
- **Invalid angles** - Only 90, 180, 270 are allowed; other values throw an error

## Metadata

The returned binary includes updated metadata:

- `width` - Image width in pixels (swapped for 90/270)
- `height` - Image height in pixels (swapped for 90/270)
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from input if present

## See Also

- [flip_image_x() - Flip horizontally](/docs/reference/flip_image_x.md)
- [flip_image_y() - Flip vertically](/docs/reference/flip_image_y.md)
- [scale_image() - Resize images](/docs/reference/scale_image.md)
- [crop_image() - Extract regions](/docs/reference/crop_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
