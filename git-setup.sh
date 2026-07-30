#!/bin/bash
set -e

echo "Setting up git hooks..."

# Compile the duso-tag binary
echo "Compiling duso-tag..."
go build -o ./bin/duso-tag ./cmd/duso-tag

# Create the pre-commit hook script
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -e

# Lint staged .du and .md files (duso lint handles both in one call)
lint_files=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(du|md)$' || true)
if [ -n "$lint_files" ]; then
  echo "Linting duso and markdown files..."
  echo "$lint_files" | xargs sh -c 'duso lint "$@" -ignore-warnings' _
fi
EOF

chmod +x .git/hooks/pre-commit

# Create the post-commit hook script
cat > .git/hooks/post-commit << 'EOF'
#!/bin/bash
exec "$(git rev-parse --show-toplevel)/bin/duso-tag"
EOF

chmod +x .git/hooks/post-commit

echo "Git setup complete!"
echo "Pre-commit hook installed in .git/hooks/"
echo "Post-commit hook installed in .git/hooks/"
echo "Binary compiled to ./bin/duso-tag"
echo "Next commit with 'feat:', 'fix:', or 'major:' prefix will auto-tag and push"
