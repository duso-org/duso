package runtime

// GoFunction is a function that can be called from Duso scripts.
// It takes named arguments as a map and returns a result or error.
type GoFunction func(args map[string]any) (any, error)
