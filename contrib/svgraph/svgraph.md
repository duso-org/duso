# svgraph Module

svgraph is a pure Duso module for generating scalable vector graphics (SVG) charts. Create beautiful, publication-quality charts directly from Duso scripts without external dependencies. Charts render to standalone SVG files that work in any browser or editor.

## Quick Start

```duso
sg = require("svgraph")

// Generate a line chart
svg = sg.render({
  type = "line",
  title = "Temperature Over Week",
  axis = ["Days", "Temperature (°F)"],
  data = [65, 68, 72, 70, 75, 78, 76]
})

save("chart.svg", svg)
```

## The render() Function

All charts are created with a single function: `render(spec)`. The function takes a specification object and returns SVG markup as a string.

### Specification Object

The `spec` parameter is an object with the following fields:

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `type` | Yes | string | Chart type: `"line"`, `"bar"`, `"hbar"`, `"scatter"`, `"bubble"`, `"area"`, `"multi"`, or `"donut"` |
| `data` | Yes | array | Chart data (format depends on type) |
| `title` | No | string | Chart title displayed at top |
| `axis` | No | array | Axis labels `[x_label, y_label]` |
| `labels` | No | array | X-axis tick labels, parallel to `data`. Thinned automatically so they don't overprint; a short array labels only the points it covers |
| `width` / `height` | No | number | Canvas size. Default 1200 × 800 |
| `margin` | No | object | `{top, right, bottom, left}`; any subset |
| `background` | No | string | Backing rect color, or `"none"` to omit it |
| `ink` | No | string | Axes, ticks and text |
| `series` | No | string or array | One color, or a palette cycled across series |
| `font_family` | No | string | |
| `font_size` | No | number | Label size; the title scales from it |
| `line_width` / `dot_size` | No | number | |
| `ticks` | No | number | Approximate y-axis tick count. Default 5 |
| `zero` | No | boolean | Force the value axis to include 0. Defaults true for `bar`, `hbar` and `area` |
| `bare` | No | boolean | Drop all chrome — no title, axes, ticks or margins. For sparklines |

### Structured data

Anywhere label/value pairs are accepted, objects work too, so callers
with structured data don't have to flatten it by hand:

```duso
sg.render({type = "bar", data = [{label = "Q1", value = 100}, {label = "Q2", value = 150}]})
```

Scatter takes `{x =, y =}` and bubble `{x =, y =, size =}` the same way.

### Chart Types

## Line Chart

Plot values as a continuous line with markers at each point.

**Data Format:** Array of numbers

**Use Cases:** Time series, trends, continuous measurements

```duso
sg = require("svgraph")

line_chart = sg.render({
  type = "line",
  title = "Stock Price",
  axis = ["Week", "Price ($)"],
  data = [100, 105, 102, 110, 115, 120, 118]
})

save("stock.svg", line_chart)
```

**Tips:**
- Single data point displays as a horizontal line with a dot
- Constant values (all the same) display as a flat line
- Automatically handles negative values
- Y-axis scale includes 10% padding for visual clarity

## Bar Chart (Vertical)

Display categorical data as vertical bars.

**Data Format:** Alternating label-value pairs: `[label1, value1, label2, value2, ...]`

**Use Cases:** Category comparisons, sales by quarter, discrete measurements

```duso
bar_chart = sg.render({
  type = "bar",
  title = "Sales by Quarter",
  axis = ["Quarter", "Sales ($)"],
  data = ["Q1", 100000, "Q2", 150000, "Q3", 120000, "Q4", 180000]
})

save("sales.svg", bar_chart)
```

**Features:**
- Handles both positive and negative values
- Bars extend from zero baseline
- Category labels appear below each bar
- Automatically scales Y-axis with 10% padding

## Horizontal Bar Chart

Display categorical data as horizontal bars (useful for many categories or long labels).

**Data Format:** Alternating label-value pairs: `[label1, value1, label2, value2, ...]`

**Use Cases:** Project progress, survey results, rankings

```duso
hbar_chart = sg.render({
  type = "hbar",
  title = "Project Progress",
  axis = ["Project", "% Complete"],
  data = ["Frontend", 85, "Backend", 75, "Database", 90, "Testing", 60, "Docs", 45]
})

save("progress.svg", hbar_chart)
```

## Scatter Plot

Display individual (x, y) points to reveal correlations and distributions.

**Data Format:** Alternating x-y pairs: `[x1, y1, x2, y2, x3, y3, ...]`

**Use Cases:** Correlation analysis, outlier detection, bivariate relationships

```duso
scatter = sg.render({
  type = "scatter",
  title = "Study Hours vs Grade",
  axis = ["Hours Studied", "Grade (%)"],
  data = [2, 55, 3, 60, 5, 75, 6, 85, 7, 82, 8, 90, 10, 95]
})

save("correlation.svg", scatter)
```

**Features:**
- Points plotted as circles
- Both axes scaled independently
- Useful for identifying patterns and outliers
- Handles both positive and negative coordinates

## Area Chart

Like a line chart, but with the area under the line filled in.

**Data Format:** Array of numbers (same as line chart)

**Use Cases:** Cumulative values, stacked measurements, total amount over time

```duso
area_chart = sg.render({
  type = "area",
  title = "Cumulative Revenue",
  axis = ["Month", "Total Revenue ($)"],
  data = [10000, 25000, 18000, 35000, 42000]
})

save("revenue.svg", area_chart)
```

**Features:**
- Filled area shows total magnitude
- Outline still visible for precise values
- Automatically includes 10% padding on axis scales

## Bubble Chart

Scatter plot where bubble size represents a third dimension.

**Data Format:** Triplets of x, y, size: `[x1, y1, size1, x2, y2, size2, ...]`

**Use Cases:** Market analysis (price vs rating vs volume), portfolio analysis, multidimensional comparisons

```duso
bubble = sg.render({
  type = "bubble",
  title = "Product Market Analysis",
  axis = ["Price ($)", "Customer Rating"],
  data = [
    10, 3.5, 100,     // Product A: $10, 3.5 stars, 100 units
    25, 4.2, 250,     // Product B: $25, 4.2 stars, 250 units
    15, 3.8, 180,     // Product C: $15, 3.8 stars, 180 units
    40, 4.7, 320      // Product D: $40, 4.7 stars, 320 units
  ]
})

save("products.svg", bubble)
```

**Features:**
- Bubble size scales from 3 to 18 pixels based on data range
- Useful for revealing patterns in 3D data
- X and Y axes auto-scale with 10% padding

## Multi-Line Chart

Plot multiple series on the same chart for comparison.

**Data Format:** Array of arrays (one per series): `[[series1_val1, series1_val2, ...], [series2_val1, ...], ...]`

**Use Cases:** Multi-stock comparison, resource usage over time, multiple metrics

```duso
multi = sg.render({
  type = "multi",
  title = "Stock Price Comparison",
  axis = ["Week", "Price ($)"],
  data = [
    [100, 105, 102, 110, 115, 120, 118],  // Company A
    [50, 52, 51, 55, 58, 62, 60],         // Company B
    [200, 198, 205, 210, 208, 215, 220]   // Company C
  ]
})

save("stocks.svg", multi)
```

**Requirements:**
- All series must have the same number of data points
- Up to 5+ series supported (colors auto-cycle)
- Each series gets its own color for easy distinction

**Color Scheme:**
- Green (#5FBB46)
- Blue (#2563eb)
- Red (#dc2626)
- Amber (#f59e0b)
- Purple (#8b5cf6)
- Cyan, Pink (additional colors available)

## Donut Chart

Pie-like chart showing proportions of a whole.

**Data Format:** Alternating label-value pairs: `[label1, value1, label2, value2, ...]`

**Use Cases:** Market share, budget allocation, traffic sources, composition analysis

```duso
donut = sg.render({
  type = "donut",
  title = "Market Share by Competitor",
  data = [
    "Company A", 35,
    "Company B", 28,
    "Company C", 22,
    "Company D", 10,
    "Others", 5
  ]
})

save("market.svg", donut)
```

### Sizing the ring

The ring is drawn as a fraction of the **shorter** canvas side, so a wide
canvas gives a small ring in a lot of empty space rather than a wide one.
`donut_size` is that fraction, as an outer diameter:

```duso
donut = sg.render({
  type = "donut",
  data = ["Desktop", 107, "Mobile", 69, "Tablet", 13],
  width = 1000,
  height = 800,
  donut_size = 0.6,           // default 0.375
  donut_label_gutter = 0.22,  // default 0.117
  donut_gap = 0               // default 3 degrees
})
```

`donut_gap` is the separation between slices, in degrees; `0` gives a
continuous ring, with adjacent slices sharing their arc endpoints inner and
outer. Slices are separated by angle only — every arc is struck about the one
chart centre, so the ring is a true annulus and the gaps are radial. Note the
gap is an angle, so it widens in pixels as `donut_size` grows: a ring that
looks lightly segmented at the default size can read as heavily split once
enlarged.

### Labels

Labels sit outside their slice on a leader line, never inside it. Text drawn
on a slice has to contrast with that slice's own color, which differs per
series and again between light and dark themes — there is no single ink that
works against all of them.

Nothing here measures text, because there are no font metrics. Where a leader
label looks like it will overrun its gutter, `textLength` is set so the
renderer fits the glyphs to the space instead of letting them run off the
canvas. The check that decides this is an estimate, so it errs toward
applying `textLength` when it wasn't strictly needed — a slightly condensed
label rather than a clipped one.

### Hover

Every slice and bar carries a `<title>` child, so browsers show a tooltip
naming the item, its value and — for donuts — its share. Pair it with CSS on
the `sg-` classes for highlight-on-hover, no script involved:

```css
.sg-slice:hover { filter: brightness(1.12); }
```

Light the hovered mark rather than dimming its neighbours: dimming reads as
the whole chart reacting to the pointer, which is distracting to move
across.

Labels sit `donut_label_gutter × width` in from each edge and are anchored
outward, so text wider than that gutter runs off the canvas and is
clipped. Names alone fit the default; names carrying values —
`"Desktop 107"` — need it raised. Between them, `donut_size` and
`donut_label_gutter` split the canvas: the ring wants the middle, the
labels want the edges, and pushing either too far collides with the other.

**Features:**
- Automatic label positioning with collision avoidance
- Pointer lines connect labels to segments
- Color-coded segments (7 colors available)
- Works with any number of categories
- Values are proportional (percentages calculated automatically)

## Working with Dynamic Data

svgraph works seamlessly with generated and transformed data:

```duso
sg = require("svgraph")

// Generate Fibonacci sequence
fib = []
a = 1
b = 1
for i = 1, 10 do
  push(fib, a)
  temp = a + b
  a = b
  b = temp
end

// Chart the generated data
chart = sg.render({
  type = "line",
  title = "Fibonacci Sequence",
  axis = ["Position", "Value"],
  data = fib
})

save("fibonacci.svg", chart)
```

## Data Transformation Example

Transform structured data into chart format:

```duso
sg = require("svgraph")

// Structured data
sales_data = [
  {month = "Jan", revenue = 50000},
  {month = "Feb", revenue = 65000},
  {month = "Mar", revenue = 58000},
  {month = "Apr", revenue = 75000},
  {month = "May", revenue = 82000}
]

// Transform to flat array for bar chart
flat_data = []
for data_point in sales_data do
  push(flat_data, data_point.month)
  push(flat_data, data_point.revenue)
end

chart = sg.render({
  type = "bar",
  title = "Monthly Revenue",
  axis = ["Month", "Revenue ($)"],
  data = flat_data
})

save("revenue.svg", chart)
```

## Batch Chart Generation

Generate multiple charts in a loop:

```duso
sg = require("svgraph")

periods = [
  {name = "Q1", values = [10, 15, 12]},
  {name = "Q2", values = [18, 22, 19]},
  {name = "Q3", values = [25, 28, 24]},
  {name = "Q4", values = [32, 35, 30]}
]

for period in periods do
  chart = sg.render({
    type = "line",
    title = "Performance: " + period.name,
    axis = ["Week", "Score"],
    data = period.values
  })
  save("period_" + lower(period.name) + ".svg", chart)
  print("Created: period_" + lower(period.name) + ".svg")
end
```

## Edge Cases & Robustness

svgraph handles edge cases gracefully:

```duso
sg = require("svgraph")

// Single data point
single = sg.render({
  type = "line",
  title = "Single Value",
  data = [42]
})

// Constant values (flat line)
flat = sg.render({
  type = "line",
  title = "Constant",
  data = [50, 50, 50, 50, 50]
})

// Large numbers (auto-scales axes)
large = sg.render({
  type = "bar",
  title = "Large Numbers",
  data = ["A", 1000000, "B", 2500000, "C", 1800000]
})

// Very small decimal numbers
small = sg.render({
  type = "scatter",
  title = "Precision",
  axis = ["X", "Y"],
  data = [0.001, 0.002, 0.003, 0.001, 0.005, 0.003]
})
```

All edge cases render correctly without special handling needed.

## SVG Output

The `render()` function returns valid SVG markup. Save it directly to a file:

```duso
sg = require("svgraph")
svg = sg.render({type = "line", data = [1, 2, 3, 4, 5]})
save("chart.svg", svg)
```

SVG files are:
- **Scalable**: Display at any size without quality loss
- **Embeddable**: Include in HTML/CSS, PDFs, presentations
- **Text-based**: Easy to inspect and modify with text editors
- **Browser-compatible**: Open in any modern browser
- **Print-friendly**: Render perfectly on paper

## Chart Dimensions

Default: **1200 × 800**, overridable with `width` and `height`.

Only the *ratio* matters for display. The SVG carries a `viewBox` and no
`width`/`height` attributes, so it scales to whatever box it's given —
which means the canvas should match the shape of the container it lands
in. Declaring a height in CSS as well as here is what produces
letterboxing; pick the shape once, on the canvas, and let CSS set width
with `height: auto`.

## Styling & Customization

Every element carries a class, so a page can restyle a whole chart in CSS
without touching the markup. Colors are emitted as presentation
attributes, which lose to any CSS rule, so a stylesheet always wins:

| Class | Element |
|-------|---------|
| `sg` | the root `<svg>` |
| `sg-bg` | backing rect |
| `sg-title` | title text |
| `sg-axis` | axis lines |
| `sg-tick` / `sg-tick-label` | tick marks and their text |
| `sg-axis-label` | axis names |
| `sg-series` / `sg-series-N` | lines; `N` is the series index |
| `sg-area` | area fill |
| `sg-point` | data point markers |
| `sg-bar` | bars |
| `sg-slice` | donut segments |
| `sg-label` / `sg-leader` / `sg-leader-dot` | donut labels and pointers |

```css
.sg-series { stroke: var(--primary); }
.sg-area   { fill: var(--primary); }
.sg-axis, .sg-tick { stroke: var(--border); }
.sg-tick-label { fill: var(--text-muted); }
```

This is the way to theme a chart on a page that has light and dark modes:
render with `background = "none"` so the card behind it shows through,
and let the tokens do the rest.

Color options (`ink`, `series`) accept `currentColor`. They do **not**
accept `var(--x)` — CSS variables don't resolve inside SVG presentation
attributes. Use the classes above for that.

## Error Handling

The module validates data and throws clear errors:

```duso
try
  // Missing data field
  sg.render({type = "line"})
catch (e)
  print("Error: " + e)  // "Error: Need: line|bar|..."
end

try
  // Wrong data format for bar chart
  sg.render({type = "bar", data = [1, 2, 3]})
catch (e)
  print("Error: " + e)  // "Error: Need label,value pairs"
end

try
  // Bubble chart needs triplets
  sg.render({type = "bubble", data = [1, 2, 3, 4]})
catch (e)
  print("Error: " + e)  // "Error: Need x,y,size triplets"
end
```

## Performance Notes

svgraph generates charts efficiently:
- Pure Duso implementation (no external dependencies)
- Single-pass rendering (fast for typical datasets)
- Scales well with typical business data (hundreds to thousands of points)
- Embedded in binary (no HTTP calls or external libraries needed)

## Examples

See the `contrib/svgraph/examples/` directory for complete, runnable examples:
- **basic.du**: All chart types with simple data
- **advanced.du**: Dynamic generation, transformation, batch processing
- **new_charts.du**: Examples of bubble, multi, and donut charts

Run any example:
```bash
duso contrib/svgraph/examples/basic.du
```

## Integration with Claude

Use svgraph in Claude-powered workflows to automatically visualize AI analysis:

```duso
claude = require("claude")
sg = require("svgraph")

// Claude analyzes data and returns numeric results
response = claude.prompt("Analyze these sales numbers...")
data = parse_json(response)

// Visualize the results
chart = sg.render({
  type = "bar",
  title = "Analysis Results",
  data = data
})

save("analysis.svg", chart)
print("Visualization saved to analysis.svg")
```

## FAQ

**Q: Can I customize colors and fonts?**
A: The module uses fixed styling for simplicity. Modify `contrib/svgraph/svgraph.du` to customize the `style()` function.

**Q: How many data points can I chart?**
A: Thousands of points work fine. For millions of points, consider aggregation.

**Q: Can I embed the SVG in HTML?**
A: Yes! Either embed the SVG directly or reference it: `<img src="chart.svg">` or `<embed src="chart.svg">`

**Q: What if my data has missing values?**
A: Pass valid numbers only. Use Claude or Duso functions to preprocess data and handle missing values before charting.

**Q: Can I animate SVG charts?**
A: SVG supports CSS animations. Edit the output SVG to add `<style>` tags with keyframes.

## See Also

- [Duso Learning Guide](/docs/learning-duso.md) - Language fundamentals
- [require() Reference](/docs/reference/require.md) - Module loading
- [save() Reference](/docs/reference/save.md) - File output
- [Claude Module](/contrib/claude/claude.md) - AI-powered analysis
