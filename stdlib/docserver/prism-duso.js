/**
 * Prism language definition for Duso
 * Based on the VSCode TextMate grammar for Duso
 */
Prism.languages.duso = {
  'comment': [
    {
      pattern: /\/\*[\s\S]*?\*\//,
      greedy: true
    },
    {
      pattern: /\/\/.*/,
      greedy: true
    }
  ],
  'string': [
    // Regex patterns with ~...~
    {
      pattern: /~(?:\\.|[^\\\n~])*~/,
      greedy: true,
      inside: {
        'escape': /\\./
      }
    },
    // Multiline double-quoted strings with template expressions
    {
      pattern: /"""(?:\\.|[\s\S])*?"""/,
      greedy: true,
      inside: {
        'template-expression': {
          pattern: /\{\{[\s\S]*?\}\}/,
          inside: {
            'punctuation': /[{}]/,
            'function': /[a-z_]\w*/,
            'operator': /[+\-*/%=<>!&|]/,
            'number': /\b\d+(?:\.\d+)?\b/
          }
        },
        'escape': /\\./
      }
    },
    // Multiline single-quoted strings with template expressions
    {
      pattern: /'''(?:\\.|[\s\S])*?'''/,
      greedy: true,
      inside: {
        'template-expression': {
          pattern: /\{\{[\s\S]*?\}\}/,
          inside: {
            'punctuation': /[{}]/,
            'function': /[a-z_]\w*/,
            'operator': /[+\-*/%=<>!&|]/,
            'number': /\b\d+(?:\.\d+)?\b/
          }
        },
        'escape': /\\./
      }
    },
    // Double-quoted strings with template expressions
    {
      pattern: /"(?:\\.|[^\\"\n])*"/,
      greedy: true,
      inside: {
        'template-expression': {
          pattern: /\{\{[^}]*?\}\}/,
          inside: {
            'punctuation': /[{}]/,
            'function': /[a-z_]\w*/,
            'operator': /[+\-*/%=<>!&|]/,
            'number': /\b\d+(?:\.\d+)?\b/
          }
        },
        'escape': /\\./
      }
    },
    // Single-quoted strings with template expressions
    {
      pattern: /'(?:\\.|[^\\'\n])*'/,
      greedy: true,
      inside: {
        'template-expression': {
          pattern: /\{\{[^}]*?\}\}/,
          inside: {
            'punctuation': /[{}]/,
            'function': /[a-z_]\w*/,
            'operator': /[+\-*/%=<>!&|]/,
            'number': /\b\d+(?:\.\d+)?\b/
          }
        },
        'escape': /\\./
      }
    }
  ],
  'keyword': [
    // Control flow keywords
    {
      pattern: /\b(?:if|then|elseif|else|end|while|do|for|in|function|return|break|continue|try|catch|var|raw)\b/
    },
    // Logical operators
    {
      pattern: /\b(?:and|or|not)\b/
    }
  ],
  'builtin': {
    pattern: /\b(?:abs|append|ceil|clamp|contains|floor|round|sqrt|pow|max|min|format_json|parse_json|format_time|parse_time|join|keys|len|load|lower|print|replace|save|sleep|sort|split|substr|tostring|tonumber|tobool|trim|type|upper|values|map|filter|reduce|breakpoint|watch|env|input|exit|throw|doc|run|spawn|parallel|context|datastore|http_client|http_server|include|require|range|sys|now)\b/
  },
  'constant': {
    pattern: /\b(?:true|false|nil)\b/
  },
  'number': {
    pattern: /\b\d+(?:\.\d+)?\b/
  },
  // Capitalized identifiers followed by ( are constructor/type calls
  'constructor': {
    pattern: /\b[A-Z][A-Za-z0-9_]*(?=\()/
  },
  // Function names followed by parentheses
  'function': /\b[a-z_]\w*(?=\s*\()/,
  'operator': [
    // Comparison operators
    {
      pattern: /(?:==|!=|<=|>=|<|>)/
    },
    // Arithmetic operators
    {
      pattern: /[+\-*/%]/
    },
    // Assignment
    {
      pattern: /=/
    },
    // Ternary
    {
      pattern: /\?/
    }
  ],
  'punctuation': /[{}[\]().,;]/
};

// Add language alias if needed
Prism.languages.du = Prism.languages.duso;
