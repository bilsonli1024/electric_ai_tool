#!/bin/bash
cd "$(dirname "$0")"
# Use Go 1.24.6
GO_BIN="/Users/bilson.li/.gvm/gos/go1.24.6/bin/go"
$GO_BIN build -o electric_ai_tool . >build.log 2>&1
exit_code=$?
echo "Exit code: $exit_code"
if [ -f build.log ]; then
    echo "=== Build log ==="
    cat build.log
fi
exit $exit_code
