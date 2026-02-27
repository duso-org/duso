# Plan: Fyne GUI Integration for Duso

## Context

Add desktop GUI support to Duso via a new `duso-ui` binary that wraps the Fyne GUI framework. Scripts use a `fyne()` builtin (mirrors `http_server()` pattern) that returns an object with all UI methods. The separate binary avoids adding CGo/system dependencies to the main `duso` binary.

## Key Constraint: Fyne Threading

On macOS, `app.Run()` **must** run on the main OS thread. This requires:
- `cmd/duso-ui/main.go` creates the Fyne app and calls `a.Run()` on the main thread
- The Duso interpreter/script runs in a goroutine
- All Fyne API calls from the script goroutine go through `fyne.DoAndWait()` (blocking) or `fyne.Do()` (fire-and-forget)
- Callbacks fired from Fyne's main thread can call Duso functions directly (no `fyne.Do` needed; or `fyne.Do` is a no-op synchronous call on the main thread)

## API Design

```duso
ui = fyne({title="My App"})   // optional config

// Windows
win = ui.new_window("My App")
win.set_content(container)
win.resize(400, 300)
win.center()
win.show()
win.close()
win.on_close(func() { ... })

// Widgets (each returns a widget object with methods)
lbl = ui.label("Hello!")
lbl.set_text("Updated")

btn = ui.button("Click me", func() { print("clicked!") })
btn.set_text("New label")

entry = ui.entry("placeholder...")
entry.set_text("initial value")
entry.on_changed(func(text) { print(text) })
text = entry.get_text()

chk = ui.check("Enable feature", func(v) { print(v) })
sel = ui.select(["A", "B", "C"], func(v) { print(v) })
slider = ui.slider(0, 100)
slider.on_changed(func(v) { print(v) })
bar = ui.progress()
bar.set_value(0.5)

// Layouts / Containers
c = ui.vbox(lbl, btn, entry)
c = ui.hbox(lbl, btn)
c = ui.border(nil, nil, nil, nil, center_widget)  // border layout
c = ui.padded(widget)
c = ui.center(widget)
c = ui.scroll(widget)
split = ui.hsplit(left, right)
split = ui.vsplit(top, bottom)
card = ui.card("Title", "Subtitle", content)
```

## Architecture

### Files to Create

**`pkg/ui/registry.go`** - Widget pointer registry
Stores Go `fyne.CanvasObject` pointers keyed by UUID string. Each Duso widget object includes a hidden `_fyne_id` field. Container builtins extract `_fyne_id` from child objects and look up the Go pointer.

```go
var widgetCounter int64
var widgetMap sync.Map

func RegisterWidget(w fyne.CanvasObject) string {
    id := fmt.Sprintf("_fyne_%d", atomic.AddInt64(&widgetCounter, 1))
    widgetMap.Store(id, w)
    return id
}
func LookupWidget(id string) (fyne.CanvasObject, bool) { ... }
func ExtractID(arg any) (string, bool) { ... }  // unwraps map[string]any → _fyne_id
```

**`pkg/ui/app.go`** - Global app reference + `fyne()` builtin
Holds the `fyne.App` instance set before `a.Run()`. The `fyne()` builtin returns the UI object with all child builtins as closures.

```go
var globalApp fyne.App
func SetGlobalApp(a fyne.App) { globalApp = a }
func builtinFyne(evaluator *Evaluator, args map[string]any) (any, error) {
    return map[string]any{
        "new_window": NewGoFunction(builtinNewWindow),
        "label":      NewGoFunction(builtinLabel),
        "button":     NewGoFunction(builtinButton),
        "entry":      NewGoFunction(builtinEntry),
        "check":      NewGoFunction(builtinCheck),
        "select":     NewGoFunction(builtinSelect),
        "slider":     NewGoFunction(builtinSlider),
        "progress":   NewGoFunction(builtinProgress),
        "vbox":       NewGoFunction(builtinVBox),
        "hbox":       NewGoFunction(builtinHBox),
        "border":     NewGoFunction(builtinBorder),
        "padded":     NewGoFunction(builtinPadded),
        "center":     NewGoFunction(builtinCenter),
        "scroll":     NewGoFunction(builtinScroll),
        "hsplit":     NewGoFunction(builtinHSplit),
        "vsplit":     NewGoFunction(builtinVSplit),
        "card":       NewGoFunction(builtinCard),
    }, nil
}
```

**`pkg/ui/widgets.go`** - Widget creation builtins
Each widget builtin:
1. Calls `fyne.DoAndWait()` to create the widget on the main thread
2. Registers it in the widget registry
3. Returns a Duso object with `_fyne_id` + method closures

```go
func builtinButton(evaluator *Evaluator, args map[string]any) (any, error) {
    label, _ := args["0"].(string)
    callbackArg := args["1"]
    callbackVal := InterfaceToValue(callbackArg)

    var btn *widget.Button
    fyne.DoAndWait(func() {
        btn = widget.NewButton(label, func() {
            // On Fyne main thread - use child evaluator for isolation
            childEval := NewChildEvaluator(evaluator)
            childEval.CallFunction(callbackVal, map[string]Value{})
        })
    })
    id := RegisterWidget(btn)

    return map[string]any{
        "_fyne_id": id,
        "set_text": NewGoFunction(func(e *Evaluator, a map[string]any) (any, error) {
            text, _ := a["0"].(string)
            fyne.Do(func() { btn.SetText(text) })
            return nil, nil
        }),
    }, nil
}
```

**`pkg/ui/containers.go`** - Container/layout builtins
Container builtins collect child widget IDs from variadic args, look up Go pointers, then create Fyne containers via `fyne.DoAndWait`.

```go
func builtinVBox(evaluator *Evaluator, args map[string]any) (any, error) {
    objects := collectChildren(args)  // extracts _fyne_id from each arg
    var c *fyne.Container
    fyne.DoAndWait(func() {
        c = container.NewVBox(objects...)
    })
    id := RegisterWidget(c)
    return containerObject(id, c), nil  // returns {_fyne_id, add, remove, refresh}
}
```

**`pkg/ui/register.go`** - Registration helper called by `cmd/duso-ui`
```go
func RegisterBuiltins(interp *script.Interpreter, a fyne.App) {
    SetGlobalApp(a)
    script.RegisterBuiltin("fyne", builtinFyne)
}
```

**`cmd/duso-ui/main.go`** - New entry point
Identical to `cmd/duso/main.go` except:
1. Imports `fyne.io/fyne/v2/app` and `pkg/ui`
2. Before running the script: creates Fyne app, calls `ui.RegisterBuiltins(interp, a)`
3. Runs script in goroutine
4. Calls `a.Run()` on main thread (blocks)

```go
func main() {
    // (same flag parsing as cmd/duso)

    a := app.New()
    ui.RegisterBuiltins(a)   // registers fyne() builtin globally

    interp, err := setupInterpreter(scriptPath)

    go func() {
        _, err := interp.Execute(string(source))
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            a.Quit()
        }
    }()

    a.Run()  // blocks main thread until all windows close
}
```

### Files to Modify

- **`go.mod`** - Add `fyne.io/fyne/v2` (will also update `go.sum`)
- **`build.sh`** - Add second build step: `go build -o bin/duso-ui ./cmd/duso-ui`

### Callback Evaluator Pattern

For Duso function values passed as callbacks (button.OnTapped, entry.OnChanged):
- Capture the Duso `Value` in Go closure
- When callback fires (on Fyne main thread), create a child evaluator: `NewChildEvaluator(evaluator)` using the parent's environment so closures still work
- Call via `childEval.CallFunction(callbackVal, args)`
- Any UI calls made from within the Duso callback use `fyne.Do()` which is synchronous on main thread - no deadlock

Need to verify if `NewChildEvaluator` or equivalent exists, or if we use the pattern from `builtin_parallel.go`:
```go
childEval := NewEvaluator()
childEnv  := NewChildEnvironment(evaluator.GetEnv())
childEval.SetEnvironment(childEnv)
```

### Widget Object Helper

Each widget/container returns an object with `_fyne_id` and method closures. A helper:
```go
func widgetObject(id string, extraMethods map[string]any) map[string]any {
    obj := map[string]any{"_fyne_id": id}
    for k, v := range extraMethods { obj[k] = v }
    return obj
}
```

## Example Script

```duso
// hello.du - run with: duso-ui hello.du
ui = fyne()
win = ui.new_window("Hello Duso")

input = ui.entry("Type your name...")
btn = ui.button("Greet", func() {
    name = input.get_text()
    print("Hello, " + name + "!")
})

content = ui.vbox(
    ui.label("What's your name?"),
    input,
    btn
)

win.set_content(content)
win.resize(400, 200)
win.center()
win.show()
```

## Verification

1. `go get fyne.io/fyne/v2` in the duso repo
2. `go build -o bin/duso-ui ./cmd/duso-ui` - confirm builds cleanly
3. Write `examples/hello_ui.du` and run `bin/duso-ui examples/hello_ui.du`
4. Verify window appears, button click fires Duso callback, `print()` shows in terminal
5. Verify entry widget: type text, callback fires with correct string
6. Verify containers: vbox stacks vertically, hbox horizontally
7. Verify window close terminates the program cleanly

## Phased Delivery

**Phase 1 (core):** `fyne()` builtin, `new_window`, `label`, `button`, `vbox`, `hbox`, `entry`, `progress`. Enough for real apps.

**Phase 2:** `check`, `select`, `slider`, `card`, `scroll`, `border`, `hsplit/vsplit`, `padded`, `center`

**Phase 3:** Canvas primitives (image from file/bytes, rectangle, text), Form widget, List widget
