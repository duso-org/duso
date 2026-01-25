**Debug Mode:**

Call `breakpoint()` in your code to pause execution when running with `-debug`. Future versions will support:
- `c` - Continue execution
- `n` - Step to next statement
- `s` - Step into function

**Environment Variables:**

- `NO_COLOR` - Disable colors globally (set to any value to disable)
- `DUSO_LIB` - Colon-separated list of directories to search for modules

## Examples

Run a script:

```
duso examples/basic.du
```

Execute inline code:

```
duso -c 'print("Hello, Duso!")'
duso -c '
names = ["Alice", "Bob", "Charlie"]
for name in names do
  print(name)
end
'
```

Start REPL (interactive mode):

```
duso -repl
duso> x = 5
duso> y = 10
duso> print(x + y)
15
duso> function add(a, b) \
    > return a + b \
    > end
duso> print(add(3, 7))
10
duso> exit()
```

Run with verbose output:

```
duso -v ../scripts/my-script.du
```

View module documentation:

```
duso -doc http
duso -doc map
duso -doc claude
```
