#!/bin/bash
set -e

echo "Setting up git hooks..."

# Compile the duso-tag binary
echo "Compiling duso-tag..."
go build -o .git/hooks/duso-tag ./cmd/duso-tag

# Create post-commit hook
echo "Installing post-commit hook..."
cat > .git/hooks/post-commit << 'EOF'
#!/bin/bash

# Auto-version and tag based on commit message
# Calls duso-tag binary in same directory as this hook

HOOK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NEW_TAG=$("$HOOK_DIR/duso-tag")

if [ $? -eq 0 ]; then
  echo "Tagged: $NEW_TAG"
else
  echo "Version update failed"
  exit 1
fi
EOF

chmod +x .git/hooks/post-commit

echo "Git setup complete!"
echo "Hooks installed in .git/hooks/"
echo "Next commit with 'feat:', 'fix:', or 'major:' prefix will auto-tag"
