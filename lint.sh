#!/bin/bash

# Set GOPATH if not set
if [ -z "$GOPATH" ]; then
    export GOPATH=$(go env GOPATH)
fi

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null && [ ! -f "$GOPATH/bin/golangci-lint" ]; then
    echo "golangci-lint is not installed. Installing..."
    # Install golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOPATH/bin" latest
    echo "golangci-lint installed successfully."
fi

echo "Running golangci-lint..."
"$GOPATH/bin/golangci-lint" run --timeout=5m

exit_code=$?
if [ $exit_code -eq 0 ]; then
    echo "✅ Lint check passed!"
else
    echo "❌ Lint check failed. Please fix the issues above."
fi

exit $exit_code
