# The Duso Language Specification

**Version 0.23 — Draft**
**March 25 2026**

© 2026 Ludonode LLC — Licensed under Apache 2.0

-----

## 1. Introduction

### 1.1 Purpose

This document is the semi-formal specification of the Duso programming language. It defines the lexical grammar, syntactic grammar (in EBNF), type system, evaluation semantics, scoping rules, concurrency model, module system, and built-in function signatures. It is intended to serve as the authoritative reference for implementors, tool authors, and embedders.

### 1.2 Scope

This specification covers the core language as implemented by the reference interpreter (v0.23.x). It does not specify the standard library or community-contributed modules beyond their interface contracts, nor does it prescribe CLI behavior except where CLI semantics affect language evaluation (e.g., module resolution paths).

### 1.3 Notation Conventions

Grammatical productions use Extended Backus–Naur Form (EBNF) with the following conventions:

|Notation   |Meaning                   |
|-----------|--------------------------|
|`=`        |Definition                |
|`          |`                         |
|`( ... )`  |Grouping                  |
|`[ ... ]`  |Optional (zero or one)    |
|`{ ... }`  |Repetition (zero or more) |
|`" ... "`  |Terminal string literal   |
|`(* ... *)`|Comment                   |
|`- `       |Exception (set difference)|

Prose sections labeled **Semantics** describe evaluation behavior. Sections labeled **Constraints** describe static or dynamic requirements that, if violated, produce an error.

### 1.4 Conformance

A conforming implementation must accept all programs that are valid according to this specification and reject all programs that are invalid. Evaluation of valid programs must produce the observable effects described herein. Behavior described as *implementation-defined* may vary; behavior described as *undefined* places no obligation on the implementation.

-----

## 2. Lexical Structure

### 2.1 Source Text

A Duso source file is a sequence of Unicode code points encoded as UTF-8. The file extension is conventionally `.du`.

```ebnf
SourceFile = { Statement } EOF ;
```

### 2.2 Line Terminators

```ebnf
LineTerminator = LF | CR | CR LF ;
```

Where `LF` is U+000A and `CR` is U+000D. Line terminators are significant only insofar as they terminate single-line comments and are whitespace between tokens.

### 2.3 Whitespace

```ebnf
Whitespace = " " | "\t" | LineTerminator ;
```

Whitespace separates tokens but is otherwise not significant. Duso is not indentation-sensitive.

### 2.4 Comments

```ebnf
SingleLineComment = "//" { character - LineTerminator } LineTerminator ;

MultiLineComment  = "/*" { character | MultiLineComment } "*/" ;
```

**Semantics.** Comments are treated as whitespace. Multi-line comments may be nested; a `/*` inside a multi-line comment opens a new nesting level that must be closed by a matching `*/`.

### 2.5 Identifiers

```ebnf
Identifier  = Letter { Letter | Digit | "_" } ;
Letter      = "a" ... "z" | "A" ... "Z" | "_" ;
Digit       = "0" ... "9" ;
```

Identifiers are case-sensitive. The following identifiers are reserved keywords and may not be used as variable names:

```
and       break     catch     continue    do
else      elseif    end       false       for
function  if        in        nil         not
or        raw       return    then        true
try       var       while
```

### 2.6 Literals

#### 2.6.1 Number Literals

```ebnf
NumberLiteral   = IntegerPart [ "." FractionalPart ] ;
IntegerPart     = Digit { Digit } ;
FractionalPart  = Digit { Digit } ;
```

All numbers are represented at runtime as IEEE 754 64-bit floating-point values (`float64`).

#### 2.6.2 String Literals

```ebnf
StringLiteral       = SingleQuoteString | DoubleQuoteString | TripleQuoteString ;

SingleQuoteString   = "'" { StringChar | EscapeSequence | TemplateExpr } "'" ;
DoubleQuoteString   = '"' { StringChar | EscapeSequence | TemplateExpr } '"' ;
TripleQuoteString   = '"""' { character | TemplateExpr } '"""' ;

TemplateExpr        = "{{" Expression "}}" ;

EscapeSequence      = "\" ( "n" | "t" | "r" | "\" | '"' | "'" | "{{" ) ;
```

**Semantics.**

- Single-quoted and double-quoted strings are semantically identical.
- Triple-quoted strings preserve embedded newlines. Leading whitespace common to all lines (determined by the closing `"""` indentation) is stripped.
- Template expressions (`{{...}}`) are evaluated at string-creation time in the enclosing scope. The result is coerced to a string via the same rules as `tostring()`.
- The `raw` keyword preceding a string literal suppresses template interpolation; `{{` and `}}` are treated as literal characters.

#### 2.6.3 Boolean Literals

```ebnf
BooleanLiteral = "true" | "false" ;
```

#### 2.6.4 Nil Literal

```ebnf
NilLiteral = "nil" ;
```

#### 2.6.5 Array Literals

```ebnf
ArrayLiteral = "[" [ ExpressionList ] "]" ;
ExpressionList = Expression { "," Expression } [ "," ] ;
```

A trailing comma is permitted.

#### 2.6.6 Object Literals

```ebnf
ObjectLiteral   = "{" [ FieldList ] "}" ;
FieldList       = Field { "," Field } [ "," ] ;
Field           = FieldKey "=" Expression ;
FieldKey        = Identifier | "[" Expression "]" ;
```

Keys are identifiers (bare words) or computed expressions in brackets. Duplicate keys are allowed; the last value wins.

#### 2.6.7 Regex Literals

```ebnf
RegexLiteral = "~" RegexBody "~" ;
RegexBody    = { character - "~" } ;
```

**Semantics.** The body is compiled as a Go `regexp` pattern (RE2 syntax). Invalid patterns produce a parse-time error. Regex values are a subtype of string at runtime; when serialized across process boundaries they become plain strings.

### 2.7 Operators and Punctuation

```
+    -    *    /    %    =
==   !=   <    >    <=   >=
(    )    [    ]    {    }
,    .    :    ?    {{   }}
```

The keywords `and`, `or`, and `not` serve as logical operators.

-----

## 3. Syntactic Grammar

### 3.1 Program Structure

```ebnf
Program     = { Statement } ;
Statement   = Assignment
            | VarDeclaration
            | IfStatement
            | WhileStatement
            | ForStatement
            | FunctionDeclaration
            | TryCatchStatement
            | ReturnStatement
            | BreakStatement
            | ContinueStatement
            | ExpressionStatement ;
```

Statements are **not** terminated by semicolons. Newlines and token boundaries provide implicit statement separation.

### 3.2 Expressions

```ebnf
Expression        = TernaryExpr ;

TernaryExpr       = OrExpr [ "?" Expression ":" Expression ] ;

OrExpr            = AndExpr { "or" AndExpr } ;
AndExpr           = NotExpr { "and" NotExpr } ;
NotExpr           = [ "not" ] ComparisonExpr ;
ComparisonExpr    = AddExpr { ( "==" | "!=" | "<" | ">" | "<=" | ">=" ) AddExpr } ;
AddExpr           = MulExpr { ( "+" | "-" ) MulExpr } ;
MulExpr           = UnaryExpr { ( "*" | "/" | "%" ) UnaryExpr } ;
UnaryExpr         = [ "-" ] PostfixExpr ;
PostfixExpr       = PrimaryExpr { Postfix } ;
Postfix           = CallExpr | IndexExpr | DotExpr ;
CallExpr          = "(" [ ArgumentList ] ")" ;
IndexExpr         = "[" Expression "]" ;
DotExpr           = "." Identifier ;
ArgumentList      = Argument { "," Argument } ;
Argument          = [ Identifier "=" ] Expression ;

PrimaryExpr       = NumberLiteral
                  | StringLiteral
                  | BooleanLiteral
                  | NilLiteral
                  | ArrayLiteral
                  | ObjectLiteral
                  | RegexLiteral
                  | FunctionExpr
                  | Identifier
                  | "(" Expression ")" ;
```

#### 3.2.1 Operator Precedence (Highest to Lowest)

|Precedence|Operators / Forms|Associativity|
|----------|-----------------|-------------|
|1         |`.` `[]` `()`    |Left         |
|2         |Unary `-`        |Right        |
|3         |`*` `/` `%`      |Left         |
|4         |`+` `-`          |Left         |
|5         |`<` `>` `<=` `>=`|Left         |
|6         |`==` `!=`        |Left         |
|7         |`not`            |Right        |
|8         |`and`            |Left         |
|9         |`or`             |Left         |
|10        |`? :`            |Right        |
|11        |`=`              |Right        |

#### 3.2.2 String Concatenation

The `+` operator, when either operand is a string, performs concatenation. The non-string operand is coerced via `tostring()`.

### 3.3 Variable Declarations and Assignment

```ebnf
VarDeclaration = "var" Identifier "=" Expression ;
Assignment     = LValue "=" Expression ;
LValue         = Identifier
               | PostfixExpr IndexExpr
               | PostfixExpr DotExpr ;
```

**Semantics.**

- `var x = expr` creates a new binding in the current (innermost) scope, shadowing any binding of the same name in enclosing scopes.
- `x = expr` (without `var`) walks up the scope chain. If `x` is found, it is mutated in the scope where it was found. If `x` is not found and the current scope is a function scope, a new local binding is created. If `x` is not found and the current scope is the global scope, a new global binding is created.

### 3.4 Control Flow

#### 3.4.1 If Statement

```ebnf
IfStatement = "if" Expression "then" { Statement }
              { "elseif" Expression "then" { Statement } }
              [ "else" { Statement } ]
              "end" ;
```

**Semantics.** The condition expression is evaluated and tested for truthiness (see §4.3). The first branch whose condition is truthy is executed; at most one branch executes.

#### 3.4.2 While Statement

```ebnf
WhileStatement = "while" Expression "do" { Statement } "end" ;
```

**Semantics.** The condition is evaluated before each iteration. If truthy, the body executes; otherwise the loop terminates. `break` exits the innermost enclosing loop. `continue` skips to the next evaluation of the condition.

#### 3.4.3 For Statement

```ebnf
ForStatement = NumericFor | IteratorFor ;

NumericFor   = "for" Identifier "=" Expression "," Expression [ "," Expression ] "do"
               { Statement }
               "end" ;

IteratorFor  = "for" Identifier "in" Expression "do"
               { Statement }
               "end" ;
```

**Semantics.**

- **Numeric for:** `for i = start, end [, step] do ... end`. The loop variable `i` is local to the loop body. The `start`, `end`, and `step` expressions are evaluated exactly once, before the first iteration. `step` defaults to `1`. The loop iterates while `i <= end` (if `step > 0`) or `i >= end` (if `step < 0`). If `step == 0`, the behavior is undefined.
- **Iterator for:** `for item in collection do ... end`. If `collection` is an array, `item` takes each element value in index order. If `collection` is an object, `item` takes each key as a string. The loop variable is local to the loop body.

#### 3.4.4 Break and Continue

```ebnf
BreakStatement    = "break" ;
ContinueStatement = "continue" ;
```

**Constraints.** `break` and `continue` must appear within the body of a `for` or `while` loop. They affect the innermost enclosing loop only.

### 3.5 Function Definitions

```ebnf
FunctionDeclaration = "function" Identifier "(" [ ParameterList ] ")" { Statement } "end" ;
FunctionExpr        = "function" "(" [ ParameterList ] ")" { Statement } "end" ;
ParameterList       = Parameter { "," Parameter } ;
Parameter           = Identifier [ "=" Expression ] ;
```

**Semantics.**

- A `FunctionDeclaration` creates a named binding in the current scope, equivalent to `name = function(...) ... end`.
- Parameters are local to the function body.
- Parameters may have default values. Default expressions are evaluated at call time in the callee’s scope if the corresponding argument is not provided.
- A function captures a reference to the environment in which it is defined (its closure). This closure is used as the parent scope when the function is called.

#### 3.5.1 Calling Convention

Functions accept both positional and named arguments. At a call site:

1. Positional arguments are bound left-to-right to the parameter list.
1. Named arguments (`name = value`) bind to the parameter with the matching name, regardless of position.
1. If a positional and named argument target the same parameter, the named argument wins.
1. Excess positional arguments are silently ignored.
1. Missing arguments without default values are bound to `nil`.

#### 3.5.2 Return

```ebnf
ReturnStatement = "return" [ Expression ] ;
```

**Semantics.** Transfers control back to the caller and delivers the expression’s value (or `nil` if omitted). Implemented internally as a control-flow signal (not an exception).

### 3.6 Try/Catch

```ebnf
TryCatchStatement = "try" { Statement } "catch" "(" Identifier ")" { Statement } "end" ;
```

**Semantics.**

- The `try` block is executed. If any statement within it raises an error (via `throw()`, a runtime error, or a propagated error from a called function), control transfers to the `catch` block.
- The identifier in `catch(e)` is bound to the thrown value. This value may be of any type: a string, object, array, number, etc.
- If no error occurs, the `catch` block is skipped.

#### 3.6.1 Throw

`throw(value)` is a built-in function, not a keyword. It raises an error with `value` as the payload. If `value` is a string, it becomes the error message. If `value` is an object or other type, it is delivered as-is to the `catch` block.

### 3.7 Constructor Calls (Object and Array)

Objects and arrays are callable. Calling them produces a shallow copy with optional modifications:

```ebnf
(* Implicit in PostfixExpr / CallExpr grammar *)
(* ObjectValue "(" [ NamedArgList ] ")" *)
(* ArrayValue  "(" [ PositionalArgList ] ")" *)
```

**Semantics.**

- **Object constructor:** `obj(field1 = val1, field2 = val2, ...)` returns a new object that is a shallow copy of `obj`, with the named fields overridden. If no arguments are given, an unmodified shallow copy is produced.
- **Array constructor:** `arr(val1, val2, ...)` returns a new array that is a shallow copy of `arr` with the positional arguments appended. If no arguments are given, an unmodified shallow copy is produced.

-----

## 4. Type System

### 4.1 Types

Duso is dynamically typed. Every runtime value belongs to exactly one of ten types:

|Type      |Go Representation               |Description                                                                                                                     |
|----------|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
|`nil`     |`nil`                           |The absence of a value.                                                                                                         |
|`number`  |`float64`                       |IEEE 754 double-precision floating point.                                                                                       |
|`string`  |`string`                        |Immutable UTF-8 character sequence.                                                                                             |
|`boolean` |`bool`                          |`true` or `false`.                                                                                                              |
|`array`   |`[]Value`                       |Ordered, 0-indexed, mutable sequence.                                                                                           |
|`object`  |`map[string]Value`              |Unordered mutable mapping of string keys.                                                                                       |
|`function`|`ScriptFunction` or `GoFunction`|Callable value with captured closure.                                                                                           |
|`code`    |`*AST`                          |Pre-parsed source code. Produced by `parse()`. Can be passed to `run()` in place of a file path.                                |
|`error`   |`*DusoError`                    |Error value carrying a message, source position, and stack trace. Produced by runtime errors or `throw()`.                      |
|`binary`  |`[]byte`                        |Raw byte sequence. Produced by `decode_base64()` and binary I/O operations. Not directly printable; must be encoded for display.|

The built-in `type(value)` returns the type name as a string: `"nil"`, `"number"`, `"string"`, `"boolean"`, `"array"`, `"object"`, `"function"`, `"code"`, `"error"`, or `"binary"`.

#### 4.1.1 The `code` Type

A `code` value is a pre-parsed AST produced by the built-in `parse(source [, metadata])` function. It represents syntactically valid Duso source that has been lexed and parsed but not yet evaluated. `code` values can be passed directly to `run()` or `spawn()` in place of a file path, avoiding repeated parsing of the same source.

`parse()` never throws; if the source contains syntax errors, it returns an `error` value instead.

`code` values are truthy, not callable, and not iterable. When coerced to a string (e.g., via `tostring()` or template interpolation), the original source text is returned.

#### 4.1.2 The `error` Type

An `error` value encapsulates an error message (string), source position (file, line, column), and a captured call stack. Errors are produced by:

- Runtime faults (type mismatches, undefined variables, out-of-bounds access, etc.).
- Explicit calls to `throw(value)`.
- `parse()` when given syntactically invalid source.

Error values are truthy. When coerced to a string, the error message is returned. Error values carry structured metadata accessible via dot notation: `err.message`, `err.file`, `err.line`, `err.column`, and `err.stack` (an array of call-frame objects).

Error values are distinct from error *propagation*. An error value can exist as a normal variable without triggering unwinding; only `throw()` and internal runtime faults initiate propagation.

#### 4.1.3 The `binary` Type

A `binary` value is a raw byte sequence with no encoding assumption. Binary values are produced by `decode_base64()` and by I/O operations that read non-text data.

Binary values are truthy (even zero-length). They cannot be concatenated with `+`, indexed with `[]`, or interpolated in templates. To use binary data as a string, it must be explicitly encoded (e.g., via `encode_base64()`). `len()` on a binary value returns the byte count.

When serialized to JSON via `format_json()`, binary values are base64-encoded as strings.

### 4.2 Type Coercion Rules

Duso performs implicit coercion only in the following contexts:

|Context                        |Rule                                         |
|-------------------------------|---------------------------------------------|
|`+` with at least one string   |Non-string operand coerced via `tostring()`. |
|Array/string index             |Index must be a number; truncated to integer.|
|Condition (`if`, `while`, `?:`)|Value tested for truthiness (§4.3).          |

All other type mismatches produce a runtime error.

### 4.3 Truthiness

The following values are **falsy**: `nil`, `false`, `0`, `""` (empty string).

All other values are **truthy**, including: empty arrays `[]`, empty objects `{}`, all function values, `code` values, `error` values, and `binary` values (even zero-length).

### 4.4 Equality

- `==` and `!=` perform structural comparison for numbers, strings, and booleans (value equality).
- For arrays, objects, functions, `code`, `error`, and `binary` values, `==` tests reference identity.
- `nil == nil` is `true`.
- Values of different types are never equal (e.g., `0 == false` is `false`; `0 == "0"` is `false`).

### 4.5 Arithmetic Operators

The operators `+` (when both operands are numbers), `-`, `*`, `/`, and `%` operate on numbers. Applying them to non-number operands (except `+` with a string) is a runtime error.

Division by zero produces `+Inf`, `-Inf`, or `NaN` per IEEE 754.

### 4.6 Short-Circuit Evaluation

`and` returns the left operand if it is falsy; otherwise evaluates and returns the right operand. `or` returns the left operand if it is truthy; otherwise evaluates and returns the right operand. Both operators return the value itself, not a coerced boolean.

-----

## 5. Scoping and Environments

### 5.1 Environment Model

A Duso environment is a linked list of frames. Each frame holds a mapping from variable names to values, a reference to its parent frame (or `nil` for the root), and a boolean flag indicating whether it is a function-scope boundary.

```
Frame { variables: Map<String, Value>, parent: Frame?, isFunctionScope: Boolean }
```

### 5.2 Scope Rules

1. **Global scope** is the root frame created when a script begins execution.
1. **Function scope** is created when a function is called. Its parent is the function’s closure (the frame in which the function was *defined*), not the frame from which it was *called*.
1. **Block scope** is created for `for` and `while` loop bodies. Loop variables are local to this scope.
1. **Parallel scope** is a function scope with an additional constraint: writes cannot propagate to the parent. The parent is read-only.

### 5.3 Variable Resolution

**Lookup** (`x`): Walk up the frame chain from the current frame. Return the first binding found.

**Assignment** (`x = v`, without `var`):

1. Walk up the frame chain. If `x` is found, mutate it in that frame.
1. If not found and the current frame is a function-scope boundary, create a new local binding in the current frame.
1. If not found and the current frame is the global scope, create a new global binding.

**Local declaration** (`var x = v`): Always create a new binding in the current frame, shadowing any existing binding of `x` in enclosing frames.

### 5.4 Object Method Binding

When a method is called via dot notation (`obj.method()`), the evaluator creates a child scope whose parent is `obj`‘s internal scope (the object’s fields are treated as variables). Within the method body, unqualified identifiers resolve first against the object’s fields, then the method’s definition closure.

This means methods can read and write sibling fields without explicit `self` or `this`.

-----

## 6. Functions

### 6.1 Script Functions

A script function consists of a name (optional for expressions), a parameter list, a body (sequence of statements), and a closure (the environment captured at definition time).

### 6.2 Go Functions (Host Functions)

Go functions registered via the embedding API have the signature:

```
func(args map[string]any) (any, error)
```

Arguments are provided as a map with keys `"0"`, `"1"`, etc. for positional arguments and the parameter name for named arguments. The returned `any` is converted to a Duso `Value`; a non-nil `error` becomes a Duso error.

### 6.3 Closures

A closure is the combination of a function and the environment in which it was defined. When a closure is invoked, the function body executes in a new frame whose parent is the captured environment — not the calling environment.

Closures capture by reference. Mutations to captured variables are visible to all closures sharing the same binding.

### 6.4 Recursion

Duso supports recursion. There is currently no enforced maximum stack depth; this is implementation-defined and may be limited in future versions.

-----

## 7. Object Model

### 7.1 Objects as Values

Objects are unordered maps from string keys to arbitrary values. Fields are accessed via dot notation (`obj.field`) or bracket notation (`obj["field"]`).

### 7.2 Objects as Constructors

Any object may be called as a function. The call `obj(k1 = v1, k2 = v2)` produces a new object that is a **shallow copy** of `obj`, with the specified fields overridden. This is Duso’s alternative to class-based instantiation.

### 7.3 Methods

A function stored as a field of an object becomes a method when invoked via dot notation. The method executes with the object’s fields in scope (see §5.4). Methods may call other methods on the same object by name.

### 7.4 No Inheritance or Prototypes

Duso does not have a prototype chain, class hierarchy, or inheritance mechanism. Code reuse is achieved through the constructor pattern (copying and overriding), closures, and module imports.

-----

## 8. Arrays

### 8.1 Indexing

Arrays are 0-indexed. `arr[i]` where `i` is a number accesses the element at index `floor(i)`. Out-of-bounds access returns `nil`.

### 8.2 Mutability

Arrays are mutable. Elements may be replaced by assignment (`arr[i] = v`), and the built-in functions `push()`, `pop()`, `shift()`, and `unshift()` modify arrays in place.

### 8.3 Arrays as Constructors

Calling an array as a function produces a shallow copy with positional arguments appended:

```
template = [1, 2, 3]
extended = template(4, 5)   // [1, 2, 3, 4, 5]
```

-----

## 9. Strings

### 9.1 Immutability

Strings are immutable sequences of UTF-8 code points. All string operations produce new strings.

### 9.2 Template Interpolation

Strings may contain template expressions delimited by `{{` and `}}`. These are evaluated eagerly at the point where the string literal is encountered. The result of each expression is coerced to a string.

The `raw` modifier suppresses interpolation:

```
raw "{{this is literal}}"   // "{{this is literal}}"
```

### 9.3 Triple-Quoted Strings

Triple-quoted strings (`"""..."""`) preserve embedded newlines. Leading common whitespace (as determined by the indentation of the closing `"""`) is stripped from each line.

### 9.4 Reusable Templates

The built-in `template(str)` function returns a function that, when called, evaluates `str`‘s template expressions in the caller’s scope at call time, rather than at definition time.

-----

## 10. Error Model

### 10.1 Error Values

Errors carry a message (string), a source position (file, line, column), and a call stack at the point of origin. The `throw()` built-in raises an error with an arbitrary payload.

### 10.2 Propagation

Unhandled errors propagate up the call stack. Each `try`/`catch` block establishes an error boundary. If an error reaches the top of the stack without being caught, the script terminates with a non-zero exit code and the error is printed to stderr.

### 10.3 Control-Flow Signals

`return`, `break`, and `continue` are implemented internally as special signal values, not user-visible errors. They cannot be caught by `try`/`catch`.

-----

## 11. Concurrency Model

### 11.1 Overview

Duso’s concurrency is built on Go goroutines. Each concurrent unit of execution has its own evaluator instance and environment chain. There is no shared mutable state between concurrent executions except through the `datastore` mechanism.

### 11.2 `parallel(fns)`

Executes an array of functions concurrently. Returns an array of results in the same order as the input.

**Semantics.**

- Each function runs in a new evaluator with a parallel scope (parent is read-only).
- All functions must complete before `parallel()` returns.
- If a function raises an error, its result is `nil`.
- The parent scope is not modified.

### 11.3 `spawn(script, context)`

Launches a script asynchronously in a background goroutine.

**Semantics.**

- Returns immediately with a numeric process ID (PID).
- The script executes in a fully isolated environment. Data passed in `context` is deep-copied.
- The spawned script accesses its context via `context()`.
- The spawned script may call `exit(value)` to produce a result.
- Functions embedded in the `context` object are stripped (replaced with `nil`) during the deep copy, because closures cannot cross scope boundaries.

### 11.4 `run(script, context [, timeout])`

Executes a script synchronously, blocking until it completes or the optional timeout (in seconds) expires.

**Semantics.**

- Data is deep-copied as with `spawn()`.
- Returns the value passed to `exit()` in the child script.
- If the timeout expires, a timeout error is raised.

### 11.5 `kill(pid)`

Terminates a previously spawned process by its PID.

### 11.6 Datastore

`datastore(namespace [, config])` returns a handle to a named, thread-safe, in-memory key-value store. Datastores provide the sole mechanism for communication between concurrent scripts.

Datastore operations include:

|Operation                           |Description                               |
|------------------------------------|------------------------------------------|
|`store.get(key)`                    |Retrieve value for key, or `nil`.         |
|`store.set(key, value)`             |Set key to value.                         |
|`store.delete(key)`                 |Remove key.                               |
|`store.increment(key, n)`           |Atomically increment numeric value.       |
|`store.push(value)`                 |Append to an internal queue.              |
|`store.pop([timeout])`              |Remove from queue tail, blocking if empty.|
|`store.shift([timeout])`            |Remove from queue head, blocking if empty.|
|`store.wait(key, value [, timeout])`|Block until key equals value, or timeout. |
|`store.keys()`                      |Return array of all keys.                 |

All operations are atomic with respect to other datastore operations on the same namespace. The optional `config` parameter `{disk = true}` enables persistence to the filesystem.

### 11.7 Serialization Contract

When data crosses a process boundary (`spawn`, `run`, `datastore`), it is deep-copied. The following rules apply:

- **Preserved:** `nil`, `number`, `string`, `boolean`, `binary`, nested `array`, nested `object`.
- **Converted:** Regex values become plain strings. `error` values become their string message. `code` values become their source string.
- **Stripped:** `function` values become `nil`.

-----

## 12. Module System

### 12.1 `require(path)`

Loads and executes a module in an isolated scope. Returns the module’s exported value (the last expression evaluated, or the value passed to `exit()`).

**Semantics.**

- The module executes in its own environment. Variables defined in the module do not leak to the caller.
- Results are cached by resolved path. Subsequent `require()` calls with the same path return the cached value without re-execution.
- Parse results (AST) are cached with mtime validation for filesystem modules; embedded modules are cached indefinitely.

### 12.2 `include(path)`

Executes a script in the **caller’s** scope. Variables defined in the included script are visible in the caller.

**Semantics.**

- No isolation; this is equivalent to textual inclusion.
- AST is cached, but the script is re-executed on every `include()` call.

### 12.3 Module Resolution Order (CLI)

When the CLI resolves a module path:

1. **Absolute paths** and paths starting with `/EMBED/` or `/STORE/` are used directly.
1. **Relative paths** (starting with `./` or `../`) are resolved relative to the currently executing script.
1. **Bare names** are searched in the following order:
   a. Directories listed in the `DUSO_LIB` environment variable.
   b. Directories specified via `-lib-path` CLI flag.
   c. Embedded standard library (`/EMBED/stdlib/`).
   d. Embedded community library (`/EMBED/contrib/`).
1. The `.du` extension is appended if not present.
1. If a directory is found, `index.du` within it is loaded.

### 12.4 Circular Dependency Detection

The module loader maintains a stack of modules currently being loaded. If a module is encountered that is already on the stack, a circular dependency error is raised, including the cycle path.

-----

## 13. Built-In Functions

The following functions are available in every scope without import. Signatures use the notation `name(param: Type [= default], ...) -> ReturnType`.

### 13.1 I/O

|Signature                          |Description                                              |
|-----------------------------------|---------------------------------------------------------|
|`print(values...: any) -> nil`     |Output values separated by spaces, followed by a newline.|
|`write(values...: any) -> nil`     |Output values without a trailing newline.                |
|`input([prompt: string]) -> string`|Read a line from stdin.                                  |

### 13.2 String Operations

|Signature                                                                                                  |Description                                            |
|-----------------------------------------------------------------------------------------------------------|-------------------------------------------------------|
|`len(value: string|array|object) -> number`                                                                |Length of string (bytes), array, or object (key count).|
|`upper(s: string) -> string`                                                                               |Uppercase conversion.                                  |
|`lower(s: string) -> string`                                                                               |Lowercase conversion.                                  |
|`trim(s: string) -> string`                                                                                |Strip leading/trailing whitespace.                     |
|`substr(s: string, pos: number [, length: number]) -> string`                                              |Substring extraction. Negative `pos` counts from end.  |
|`split(s: string, sep: string) -> array`                                                                   |Split into array. Empty `sep` splits into characters.  |
|`join(arr: array, sep: string) -> string`                                                                  |Join array elements with separator.                    |
|`contains(s: string, pattern: string|regex [, ignoreCase: boolean]) -> boolean`                            |Test if string contains pattern.                       |
|`find(s: string, pattern: string|regex) -> array`                                                          |All matches as `{text, pos, len}` objects.             |
|`replace(s: string, pattern: string|regex, replacement: string|function [, ignoreCase: boolean]) -> string`|Replace all occurrences.                               |
|`repeat(s: string, count: number) -> string`                                                               |Repeat string.                                         |
|`template(s: string) -> function`                                                                          |Create reusable template function.                     |

### 13.3 Array & Object Operations

|Signature                                                    |Description                                                         |
|-------------------------------------------------------------|--------------------------------------------------------------------|
|`push(arr: array, values...: any) -> number`                 |Append elements. Returns new length. Mutates `arr`.                 |
|`pop(arr: array) -> any`                                     |Remove and return last element. Mutates `arr`.                      |
|`shift(arr: array) -> any`                                   |Remove and return first element. Mutates `arr`.                     |
|`unshift(arr: array, values...: any) -> number`              |Prepend elements. Returns new length. Mutates `arr`.                |
|`sort(arr: array [, cmp: function]) -> array`                |Sort in place. `cmp(a, b)` returns truthy if `a` should precede `b`.|
|`map(arr: array, fn: function) -> array`                     |New array with `fn` applied to each element.                        |
|`filter(arr: array, fn: function) -> array`                  |New array of elements where `fn` returns truthy.                    |
|`reduce(arr: array, fn: function, init: any) -> any`         |Fold. `fn(accumulator, element)`.                                   |
|`keys(obj: object) -> array`                                 |Array of key strings.                                               |
|`values(obj: object) -> array`                               |Array of values.                                                    |
|`range(start: number, end: number [, step: number]) -> array`|Generate numeric sequence (inclusive).                              |
|`deep_copy(value: any) -> any`                               |Deep copy. Functions are stripped (become `nil`).                   |

### 13.4 Type Conversion

|Signature                           |Description                          |
|------------------------------------|-------------------------------------|
|`type(value: any) -> string`        |Type name.                           |
|`tonumber(value: any) -> number|nil`|Parse to number, or `nil` on failure.|
|`tostring(value: any) -> string`    |String representation.               |
|`tobool(value: any) -> boolean`     |Truthiness coercion.                 |

### 13.5 Math

|Signature                       |Description                                          |
|--------------------------------|-----------------------------------------------------|
|`abs(n) -> number`              |Absolute value.                                      |
|`floor(n) -> number`            |Round toward negative infinity.                      |
|`ceil(n) -> number`             |Round toward positive infinity.                      |
|`round(n) -> number`            |Round to nearest integer (ties to even).             |
|`sqrt(n) -> number`             |Square root.                                         |
|`pow(base, exp) -> number`      |Exponentiation.                                      |
|`min(values...) -> number`      |Minimum.                                             |
|`max(values...) -> number`      |Maximum.                                             |
|`clamp(val, min, max) -> number`|Constrain to range.                                  |
|`random() -> number`            |Pseudo-random float in [0, 1). Seeded per invocation.|
|`pi() -> number`                |The constant π.                                      |
|`sin(x)`, `cos(x)`, `tan(x)`    |Trigonometric functions (radians).                   |
|`asin(x)`, `acos(x)`, `atan(x)` |Inverse trigonometric functions.                     |
|`atan2(y, x) -> number`         |Two-argument arctangent.                             |
|`exp(x) -> number`              |e^x.                                                 |
|`log(x) -> number`              |Log base 10.                                         |
|`ln(x) -> number`               |Natural logarithm.                                   |

### 13.6 File I/O

|Signature                          |Description                             |
|-----------------------------------|----------------------------------------|
|`load(path) -> string`             |Read entire file as string.             |
|`save(path, content) -> nil`       |Write string to file (create/overwrite).|
|`append_file(path, content) -> nil`|Append to file.                         |
|`copy_file(src, dst) -> nil`       |Copy file.                              |
|`move_file(src, dst) -> nil`       |Move file.                              |
|`rename_file(old, new) -> nil`     |Rename file.                            |
|`remove_file(path) -> nil`         |Delete file.                            |
|`file_exists(path) -> boolean`     |Check existence.                        |
|`file_type(path) -> string`        |`"file"` or `"directory"`.              |
|`list_dir(path) -> array`          |Array of `{name, is_dir}` objects.      |
|`list_files(path) -> array`        |Recursive file listing.                 |
|`make_dir(path) -> nil`            |Create directory (recursive).           |
|`remove_dir(path) -> nil`          |Remove empty directory.                 |
|`current_dir() -> string`          |Working directory.                      |
|`watch(path [, timeout]) -> object`|Monitor file/directory for changes.     |

### 13.7 Encoding

|Signature                                     |Description                                                      |
|----------------------------------------------|-----------------------------------------------------------------|
|`format_json(value [, indent]) -> string`     |Serialize to JSON. Functions, errors, and binary are stringified.|
|`parse_json(s: string) -> any`                |Deserialize JSON string.                                         |
|`encode_base64(data: string|binary) -> string`|Base64 encode string or binary data.                             |
|`decode_base64(s: string) -> binary`          |Base64 decode to binary value.                                   |
|`markdown_html(text [, options]) -> string`   |Render Markdown to HTML.                                         |
|`markdown_ansi(text [, theme]) -> string`     |Render Markdown to ANSI terminal output.                         |

### 13.8 Date & Time

|Signature                           |Description                                            |
|------------------------------------|-------------------------------------------------------|
|`now() -> number`                   |Current Unix timestamp (seconds, fractional).          |
|`format_time(ts, format) -> string` |Format timestamp. Supports `"iso"` and custom patterns.|
|`parse_time(s [, format]) -> number`|Parse time string to timestamp.                        |
|`sleep(seconds) -> nil`             |Pause execution. Defaults to 1 second.                 |

### 13.9 Network

|Signature                         |Description                                                 |
|----------------------------------|------------------------------------------------------------|
|`fetch(url [, options]) -> object`|HTTP request. Returns `{ok, status, headers, body, json()}`.|
|`http_server(config) -> object`   |Create HTTP server with `.route()`, `.start()`, `.stop()`.  |

### 13.10 Security

|Signature                                   |Description                                                    |
|--------------------------------------------|---------------------------------------------------------------|
|`hash(algo, data) -> string`                |Cryptographic hash (`"sha256"`, `"sha512"`, `"sha1"`, `"md5"`).|
|`hash_password(password [, cost]) -> string`|Bcrypt hash.                                                   |
|`verify_password(password, hash) -> boolean`|Bcrypt verification.                                           |
|`sign_rsa(data, pem) -> string`             |RSA SHA256-PKCS1v15 signature.                                 |
|`verify_rsa(data, sig, pem) -> boolean`     |RSA signature verification.                                    |

### 13.11 System & Control Flow

|Signature                                         |Description                                                                        |
|--------------------------------------------------|-----------------------------------------------------------------------------------|
|`exit([value]) -> never`                          |Terminate script, returning `value`.                                               |
|`throw(value) -> never`                           |Raise error with `value` as payload.                                               |
|`context() -> object|nil`                         |Current execution context (for handlers/spawned scripts).                          |
|`parse(source [, metadata]) -> code|error`        |Parse source to AST value. Never throws.                                           |
|`run(script: string|code, ctx [, timeout]) -> any`|Execute script synchronously. Accepts a file path or a `code` value from `parse()`.|
|`spawn(script: string|code, ctx) -> number`       |Execute script asynchronously. Returns PID. Accepts a file path or a `code` value. |
|`kill(pid) -> nil`                                |Terminate spawned process.                                                         |
|`parallel(fns: array) -> array`                   |Execute functions concurrently.                                                    |
|`env(name) -> string|nil`                         |Read environment variable.                                                         |
|`uuid() -> string`                                |Generate RFC 9562 UUID v7.                                                         |
|`datastore(ns [, config]) -> object`              |Access named key-value store.                                                      |
|`doc(topic) -> string`                            |Access embedded documentation.                                                     |

### 13.12 Debugging

|Signature                       |Description                                         |
|--------------------------------|----------------------------------------------------|
|`breakpoint(values...) -> nil`  |Pause execution (requires `-debug`).                |
|`watch(exprs...: string) -> nil`|Monitor expressions for changes (requires `-debug`).|

-----

## 14. HTTP Server Model

### 14.1 Server Lifecycle

`http_server(config)` creates a server object. `config` accepts at minimum `{port: number}` and optionally TLS, CORS, JWT, and session configuration.

### 14.2 Routing

`server.route(method, pattern [, handler])` registers a handler for the given HTTP method and URL pattern. Patterns support `:param` segments for path parameters.

If `handler` is a string (file path), each incoming request spawns a new instance of that script to handle it. If `handler` is omitted, the current script handles the route (self-referential pattern, using the gate pattern with `context()`).

### 14.3 Request and Response

Within a handler script:

- `context()` returns the execution context.
- `ctx.request()` returns `{method, path, params, query, headers, body}`.
- `ctx.response()` returns an object with methods: `.json(data, status)`, `.html(body)`, `.text(body)`, `.error(status)`, `.redirect(url)`, and `.header(name, value)`.

-----

## 15. Virtual Filesystems

### 15.1 `/EMBED/`

A read-only virtual filesystem baked into the binary at build time. Contains standard library modules, community modules, documentation, and example scripts.

### 15.2 `/STORE/`

A read-write virtual filesystem backed by the datastore. Files written here persist as long as the backing datastore does (in-memory or disk-persistent).

### 15.3 Sandbox Mode

When the CLI flag `-no-files` is active, all real filesystem access is disabled. Scripts may only access `/EMBED/` (read) and `/STORE/` (read-write). Environment variable access via `env()` is also blocked.

-----

## 16. Embedding API

### 16.1 Interpreter Creation

```go
interp := script.NewInterpreter(verbose bool)
```

### 16.2 Script Execution

```go
output, err := interp.Execute(source string) (string, error)
value, err := interp.ExecuteModule(source string) (Value, error)
```

### 16.3 Registering Host Functions

```go
interp.RegisterFunction("myFunc", func(args map[string]any) (any, error) {
    // implementation
})
```

### 16.4 Configuration

```go
interp.SetDebugMode(enabled bool)
interp.SetScriptDir(dir string)
interp.SetFilePath(path string)
```

### 16.5 Adding Runtime Features

```go
cli.RegisterFunctions(interp, cli.RegisterOptions{
    ScriptDir: ".",
    HTTPPort:  8080,
})
```

This adds all file I/O, HTTP, datastore, module resolution, and concurrency functions.

-----

## 17. Execution Model Summary

1. **Lexing:** Source text → token stream.
1. **Parsing:** Token stream → AST (recursive descent, operator-precedence climbing).
1. **Evaluation:** AST → values and side effects (tree-walking interpreter).
1. **Concurrency:** Each goroutine gets its own evaluator instance. Shared state is mediated exclusively by datastores.
1. **Error handling:** Errors carry source position and call stack. Control-flow signals (`return`, `break`, `continue`) are internal and not catchable.

-----

## Appendix A: Grammar Summary

The complete EBNF grammar, consolidated:

```ebnf
Program         = { Statement } EOF ;

Statement       = VarDeclaration
                | Assignment
                | IfStatement
                | WhileStatement
                | ForStatement
                | FunctionDeclaration
                | TryCatchStatement
                | ReturnStatement
                | BreakStatement
                | ContinueStatement
                | ExpressionStatement ;

VarDeclaration  = "var" Identifier "=" Expression ;
Assignment      = LValue "=" Expression ;
LValue          = Identifier | PostfixExpr ( IndexExpr | DotExpr ) ;

IfStatement     = "if" Expression "then" { Statement }
                  { "elseif" Expression "then" { Statement } }
                  [ "else" { Statement } ]
                  "end" ;

WhileStatement  = "while" Expression "do" { Statement } "end" ;

ForStatement    = NumericFor | IteratorFor ;
NumericFor      = "for" Identifier "=" Expression "," Expression
                  [ "," Expression ] "do" { Statement } "end" ;
IteratorFor     = "for" Identifier "in" Expression "do" { Statement } "end" ;

FunctionDecl    = "function" Identifier "(" [ ParameterList ] ")"
                  { Statement } "end" ;
FunctionExpr    = "function" "(" [ ParameterList ] ")" { Statement } "end" ;
ParameterList   = Parameter { "," Parameter } ;
Parameter       = Identifier [ "=" Expression ] ;

TryCatchStmt    = "try" { Statement }
                  "catch" "(" Identifier ")" { Statement } "end" ;

ReturnStatement = "return" [ Expression ] ;
BreakStatement  = "break" ;
ContinueStatement = "continue" ;

ExpressionStmt  = Expression ;

Expression      = TernaryExpr ;
TernaryExpr     = OrExpr [ "?" Expression ":" Expression ] ;
OrExpr          = AndExpr { "or" AndExpr } ;
AndExpr         = NotExpr { "and" NotExpr } ;
NotExpr         = [ "not" ] ComparisonExpr ;
ComparisonExpr  = AddExpr { CompOp AddExpr } ;
CompOp          = "==" | "!=" | "<" | ">" | "<=" | ">=" ;
AddExpr         = MulExpr { ( "+" | "-" ) MulExpr } ;
MulExpr         = UnaryExpr { ( "*" | "/" | "%" ) UnaryExpr } ;
UnaryExpr       = [ "-" ] PostfixExpr ;
PostfixExpr     = PrimaryExpr { CallExpr | IndexExpr | DotExpr } ;
CallExpr        = "(" [ ArgumentList ] ")" ;
IndexExpr       = "[" Expression "]" ;
DotExpr         = "." Identifier ;
ArgumentList    = Argument { "," Argument } ;
Argument        = [ Identifier "=" ] Expression ;

PrimaryExpr     = NumberLiteral | StringLiteral | BooleanLiteral
                | NilLiteral | ArrayLiteral | ObjectLiteral
                | RegexLiteral | FunctionExpr | Identifier
                | "(" Expression ")" ;

ArrayLiteral    = "[" [ ExpressionList ] "]" ;
ObjectLiteral   = "{" [ FieldList ] "}" ;
FieldList       = Field { "," Field } [ "," ] ;
Field           = ( Identifier | "[" Expression "]" ) "=" Expression ;
ExpressionList  = Expression { "," Expression } [ "," ] ;
RegexLiteral    = "~" RegexBody "~" ;
```

-----

## Appendix B: Reserved Words

```
and       break     catch     continue    do
else      elseif    end       false       for
function  if        in        nil         not
or        raw       return    then        true
try       var       while
```

-----

## Appendix C: Differences from Lua

Duso’s syntax is influenced by Lua but differs in several important ways:

|Feature         |Lua                             |Duso                                        |
|----------------|--------------------------------|--------------------------------------------|
|Arrays          |1-indexed                       |0-indexed                                   |
|String templates|Not built-in                    |`{{expr}}` interpolation                    |
|Comments        |`--` and `--[[ ]]`              |`//` and `/* */` (C-style)                  |
|Regex           |Via patterns or libraries       |First-class `~regex~` literals              |
|Concurrency     |Coroutines                      |Goroutine-based (`parallel`, `spawn`, `run`)|
|Object model    |Metatables + prototype chain    |Constructor-copy pattern, no inheritance    |
|Method binding  |Explicit `self` via `:` syntax  |Implicit via object scope                   |
|`local` keyword |`local`                         |`var`                                       |
|Truthiness      |Only `nil` and `false` are falsy|`nil`, `false`, `0`, and `""` are falsy     |
|Ternary         |Not built-in                    |`condition ? a : b`                         |
