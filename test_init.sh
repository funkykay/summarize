#!/usr/bin/env bash
set -euo pipefail

ROOT="test_project"

echo "Initializing test project structure..."

# Base structure
mkdir -p "${ROOT}/docs"
mkdir -p "${ROOT}/src"

# Excluded structures (should not appear in summarize output)
mkdir -p "${ROOT}/__pycache__"
mkdir -p "${ROOT}/node_modules"
mkdir -p "${ROOT}/docs/node_modules"

# Root summarize.json for test_project
if [ ! -f "./summarize.json" ]; then
cat > "./summarize.json" <<'EOF'
{
  "excludes": [
    "__pycache__/",
    ".git/"
  ]
}
EOF
fi

# Root summarize.json for test_project
if [ ! -f "${ROOT}/summarize.json" ]; then
cat > "${ROOT}/summarize.json" <<'EOF'
{
  "excludes": [
    "node_modules"
  ]
}
EOF
fi

# Included file (should appear)
if [ ! -f "${ROOT}/docs/success.md" ]; then
cat > "${ROOT}/docs/success.md" <<'EOF'
Nice!
EOF
fi

# Included file (should appear)
if [ ! -f "${ROOT}/src/main.py" ]; then
cat > "${ROOT}/src/main.py" <<'EOF'
print("perfekt")
EOF
fi

# Excluded file (should NOT appear)
if [ ! -f "${ROOT}/__pycache__/this_should_not_be_in_summarize_result.txt" ]; then
cat > "${ROOT}/__pycache__/this_should_not_be_in_summarize_result.txt" <<'EOF'
if you see this, something is wrong
EOF
fi

# Excluded file (should NOT appear)
if [ ! -f "${ROOT}/node_modules/should_not_be_in_summarize_result.txt" ]; then
cat > "${ROOT}/node_modules/should_not_be_in_summarize_result.txt" <<'EOF'
if you see this, something is wrong
EOF
fi

# Excluded file (should NOT appear)
if [ ! -f "${ROOT}/docs/node_modules/should_not_be_in_summarize_result.txt" ]; then
cat > "${ROOT}/docs/node_modules/should_not_be_in_summarize_result.txt" <<'EOF'
if you see this, something is wrong
EOF
fi

echo "Done."
