# phospher

Inline icon fetcher module for Phospher Icons. Returns SVG
for a Phospher icon based on name and style.

## Use

```duso
ph = require("phospher")
ph.icon(name, [style], [class], [fill])
```

## Example

```duso
ph = require("phospher")
atom_svg = ph.icon("atom")
github_svg = ph.icon("github-logo", "duotone")
blue_icon = ph.icon("bell", "bold", "my-icon", "blue")
```

## API Reference

### icon(name, [style], [class], [fill])

Returns SVG string for the requested Phospher icon.

**Parameters:**
- `name` (string, required) - Icon name
- `style` (string, optional) - Icon style. Defaults to `cfg_default_style` or `"bold"` if not configured
- `class` (string, optional) - CSS class to add to SVG element. Default: `"icon-inline"`
- `fill` (string, optional) - SVG fill color. Default: `"currentColor"`

**Returns:**
- `string` - SVG markup for the icon with injected class and fill attributes, or placeholder SVG if icon not found

## Styles

- `thin`, `light`, `regular`, `bold`, `filled`, `duotone`
- defaults to `bold`

## Config

Set defaults in the `phospher.du` datastore. All optional.

```duso
ph_ds = datastore("phospher.du")

ph_ds.set("cfg_local", "path to store / load cached icon files")
ph_ds.set("cfg_default_style", "icon style to default to")
ph_ds.set("cfg_class", "CSS class to inject into SVG elements")
ph_ds.set("cfg_fill", "SVG fill color to inject into elements")
ph_ds.set("cfg_placeholder", "svg source for icons not found by name")
ph_ds.set("cfg_cdn", "url for icon CDN source")
```

## Caching

The module supports caching. Be aware if you bundle inside your app
you will want to define a true local path to use for storing new icons.
Otherwise it will default to the bundled folder. This works for loading
but not for saving since the local vfs of a bundled app is read-only.

- datastore("phospher.du") in memory
- load("path to icons folder") from local filesystem
- fetch("from phospher CDN") if not already cached somewhere

## Attribution

Icons provided by [Phosphor Icons](https://phosphoricons.com/) — a free, extensive icon set under the MIT License.

Source: [phosphor-icons/core](https://github.com/phosphor-icons/core)
