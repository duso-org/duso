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

Default chart dimensions: **1200 × 800 pixels**

The module is designed for standard web/print dimensions. Charts render with:
- **Margins**: Automatic spacing for titles, labels, and axes
- **Responsive design**: Content scales proportionally
- **High readability**: Font sizes and line widths calibrated for clarity

## Styling & Customization

The module includes built-in styling with:
- **Color scheme**: Professional, accessible colors
- **Typography**: Clean sans-serif fonts (Noto Sans, Lato, Helvetica, Arial)
- **Spacing**: Optimized margins and padding
- **Grid lines**: Y-axis tick marks and labels for easy value reading

Current styling is fixed per chart type. For custom colors or dimensions, edit the source SVG after rendering or modify the module directly.

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
