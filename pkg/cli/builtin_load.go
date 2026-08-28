package cli

import (
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// builtinLoad reads a file and returns its contents as a string.
//
// load(filename) resolves the path via ResolvePath:
//   - bare paths → appDir (entry script's directory)
//   - /HERE/...  → directory of the file the path is written in
//   - /CWD/...   → process working directory
//   - /EMBED/..., /STORE/..., absolute paths → as-is
//
// Example:
//
//	content = load("data.txt")
//	data = parse_json(load("config.json"))
//	code = load("/STORE/generated.du")
func builtinLoad(evaluator *script.Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("load() requires a filename argument")
		}
	}

	resolved := ResolvePath(filename)
	content, err := readFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("cannot load '%s': %s", filename, describeFileError(err, resolved))
	}
	return string(content), nil
}

// builtinSave writes a string to a file.
//
// save(filename, content) resolves the path via ResolvePath:
//   - bare paths → appDir (entry script's directory)
//   - /HERE/...  → directory of the file the path is written in
//   - /CWD/...   → process working directory
//   - /EMBED/... → rejected (read-only)
//   - /STORE/..., absolute paths → as-is
//
// Example:
//
//	save("output.txt", "Hello, World!")
//	save("data.json", format_json(myObject))
//	save("/STORE/generated.du", code)
func builtinSave(evaluator *script.Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("save() requires filename and content arguments")
		}
	}

	content, ok := args["1"].(string)
	if !ok {
		if c, ok := args["content"]; ok {
			content = fmt.Sprintf("%v", c)
		} else {
			return nil, fmt.Errorf("save() requires filename and content arguments")
		}
	}

	resolved := ResolvePath(filename)
	if err := writeFile(resolved, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("cannot save '%s': %s", filename, describeFileError(err, resolved))
	}

	return nil, nil
}

// builtinLoadBinary reads a binary file and returns a binary Value.
//
// load_binary(filename) uses the same resolution as load(); returns a binary
// value with metadata including the filename.
//
// Example:
//
//	image = load_binary("avatar.png")
//	save_binary(image, "backup.png")
func builtinLoadBinary(evaluator *script.Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("load_binary() requires a filename argument")
		}
	}

	resolved := ResolvePath(filename)
	content, err := readFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("cannot load_binary '%s': %s", filename, describeFileError(err, resolved))
	}
	bin := script.NewBinary(content)
	binVal := bin.AsBinary()
	binVal.Metadata["filename"] = script.NewString(filename)
	return script.InterfaceToValue(bin), nil
}

// builtinSaveBinary writes binary data to a file.
//
// save_binary(binary, filename) uses the same resolution as save().
//
// Example:
//
//	save_binary(image, "output.png")
//	save_binary(uploaded_file, "/STORE/uploads/" + filename)
func builtinSaveBinary(evaluator *script.Evaluator, args map[string]any) (any, error) {
	var binary *script.BinaryValue
	var filename string

	// Handle positional or named arguments
	if b, ok := args["0"]; ok {
		if bv, ok := b.(*script.BinaryValue); ok {
			binary = bv
		} else if bv, ok := b.(*script.ValueRef); ok && bv.Val.IsBinary() {
			binary = bv.Val.AsBinary()
		} else if bv, ok := b.(script.Value); ok && bv.IsBinary() {
			binary = bv.AsBinary()
		}
	} else if b, ok := args["binary"]; ok {
		if bv, ok := b.(*script.BinaryValue); ok {
			binary = bv
		} else if bv, ok := b.(*script.ValueRef); ok && bv.Val.IsBinary() {
			binary = bv.Val.AsBinary()
		} else if bv, ok := b.(script.Value); ok && bv.IsBinary() {
			binary = bv.AsBinary()
		}
	}

	if binary == nil {
		return nil, fmt.Errorf("save_binary() requires a binary value as first argument")
	}

	// Get filename (can be second positional or named)
	if fn, ok := args["1"].(string); ok {
		filename = fn
	} else if fn, ok := args["filename"].(string); ok {
		filename = fn
	} else if fn, ok := args["1"]; ok {
		filename = fmt.Sprintf("%v", fn)
	} else if fn, ok := args["filename"]; ok {
		filename = fmt.Sprintf("%v", fn)
	} else {
		return nil, fmt.Errorf("save_binary() requires filename argument")
	}

	resolved := ResolvePath(filename)
	if err := writeFile(resolved, *binary.Data, 0644); err != nil {
		return nil, fmt.Errorf("cannot save_binary '%s': %s", filename, describeFileError(err, resolved))
	}

	return nil, nil
}
