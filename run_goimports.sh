#!/bin/bash

# Set GOPATH if not set
if [ -z "$GOPATH" ]; then
    export GOPATH=$(go env GOPATH)
fi

# Install goimports if not already installed
if [ ! -f "$GOPATH/bin/goimports" ]; then
    echo "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

# Find all Go files and run goimports on them
echo "Running goimports on all Go files..."
find . -name "*.go" -not -path "./vendor/*" -exec "$GOPATH/bin/goimports" -w {} \;

echo "Done!"
