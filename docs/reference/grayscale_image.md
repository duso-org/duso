# grayscale_image

Converts an image to grayscale while preserving the alpha channel.

## Syntax

```duso
grayscale_image(image)
```

## Parameters

- `image` - A `binary` value containing image data (PNG, JPEG, or GIF)

## Returns

A new `binary` value containing the grayscale image with metadata preserved (width, height, format, content_type).

## Description

Converts a color image to grayscale using the standard luminosity formula. The alpha channel is preserved for PNG and GIF formats, allowing for transparent grayscale images.

## Format Support

Supports PNG, JPEG, and GIF formats. Output format matches input format. Alpha channel is preserved in PNG and GIF; JPEG results have no alpha.

## Examples

### Convert to grayscale

```duso
image = load_image("photo.jpg")
gray = grayscale_image(image)
save_image(gray, "grayscale.jpg")
```

### Grayscale with transparency

```duso
icon = load_image("icon.png")
gray_icon = grayscale_image(icon)
save_image(gray_icon, "gray_icon.png")
```

### Create grayscale thumbnail

```duso
image = load_image("photo.jpg")
result = grayscale_image(image)
  |> scale_image(200, 200, "fit")
  |> save_image("thumb_bw.jpg")
```

## Behavior

- **Luminosity formula** - Uses standard RGB to grayscale conversion
- **Preserves transparency** - Alpha channel is maintained in PNG/GIF
- **Dimensions unchanged** - Width and height remain the same

## Metadata

The returned binary includes metadata:

- `width` - Image width in pixels (unchanged)
- `height` - Image height in pixels (unchanged)
- `format` - Image format ("png", "jpeg", or "gif")
- `content_type` - MIME type ("image/png", "image/jpeg", or "image/gif")
- `filename` - Preserved from input if present

## See Also

- [convert_image() - Convert image formats](/docs/reference/convert_image.md)
- [set_image_opacity() - Adjust opacity](/docs/reference/set_image_opacity.md)
- [scale_image() - Resize images](/docs/reference/scale_image.md)
- [binary - Binary data type overview](/docs/reference/binary.md)
