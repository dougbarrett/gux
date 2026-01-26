#!/bin/bash
# Verifies no references to removed packages (components/, auth/, storage/, state/, ws/)
# Exit 0 = clean, Exit 1 = stale references found
# Excludes .planning/ directory (contains historical references by design)

set -e

DEAD_PACKAGES="components|auth|storage|state|ws"
EXIT_CODE=0

echo "=== Checking for stale package references ==="
echo ""

# Check 1: Import statements
echo "Checking import statements..."
if grep -rn "github\.com/dougbarrett/gux/\(${DEAD_PACKAGES}\)" \
    --include="*.go" --include="*.md" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" \
    --exclude-dir="vendor" . 2>/dev/null; then
    echo "FOUND: Stale import statements"
    EXIT_CODE=1
else
    echo "OK: No stale imports"
fi
echo ""

# Check 2: Package prefix usage in markdown
echo "Checking package usage in docs..."
if grep -rn "\(components\|auth\|state\|ws\)\.\(New\|Init\|Get\|Set\|Toast\|Button\|Store\|Client\|Token\)" \
    --include="*.md" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" . 2>/dev/null; then
    echo "FOUND: Stale package usage"
    EXIT_CODE=1
else
    echo "OK: No stale package usage"
fi
echo ""

# Check 3: Documentation links to deleted files
echo "Checking documentation links..."
if grep -rn "\(websocket\.md\|auth\.md\|components\.md\|state-management\.md\)" \
    --include="*.md" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" . 2>/dev/null; then
    echo "FOUND: Stale documentation links"
    EXIT_CODE=1
else
    echo "OK: No stale doc links"
fi
echo ""

# Check 4: Project structure listings
echo "Checking project structure diagrams..."
if grep -rn "├── \(components/\|auth/\|storage/\|state/\|ws/\)" \
    --include="*.md" --include="*.txt" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" . 2>/dev/null; then
    echo "FOUND: Stale structure entries"
    EXIT_CODE=1
else
    echo "OK: No stale structure entries"
fi
echo ""

# Summary
echo "=== Summary ==="
if [ $EXIT_CODE -eq 0 ]; then
    echo "SUCCESS: No stale references to removed packages found"
else
    echo "FAILURE: Stale references found - please clean up"
    echo "Packages removed: components/, auth/, storage/, state/, ws/"
fi

exit $EXIT_CODE
