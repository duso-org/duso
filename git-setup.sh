#!/bin/bash
set -e

echo "Setting up git hooks..."

# Compile the duso-tag binary
echo "Compiling duso-tag..."
go build -o ./bin/duso-tag ./cmd/duso-tag

# Create the post-commit hook script
cat > .git/hooks/post-commit << 'EOF'
#!/bin/bash
exec "$(git rev-parse --show-toplevel)/bin/duso-tag"
EOF

chmod +x .git/hooks/post-commit

echo "Git setup complete!"
echo "Post-commit hook installed in .git/hooks/"
echo "Binary compiled to ./bin/duso-tag"
echo "Next commit with 'feat:', 'fix:', or 'major:' prefix will auto-tag and push"
