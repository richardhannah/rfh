#!/bin/bash

echo "Installing CLI..."

# Create dist folder if it doesn't exist
if [ ! -d "dist" ]; then
    echo "Creating dist folder..."
    mkdir -p dist
fi

# Determine the executable name based on the OS
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    EXEC_NAME="rfh.exe"
else
    EXEC_NAME="rfh"
fi

# Build the CLI executable
echo "Building CLI executable..."
go build -o "dist/$EXEC_NAME" cmd/cli/main.go

if [ $? -ne 0 ]; then
    echo "Failed to build CLI executable"
    exit 1
fi

echo "CLI executable built successfully at dist/$EXEC_NAME"

# Get the full path to the dist directory
DIST_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/dist"

# Check if dist path is already in PATH
if [[ ":$PATH:" == *":$DIST_PATH:"* ]]; then
    echo "Dist folder is already in PATH"
else
    echo "Adding dist folder to PATH..."
    
    # Determine which shell profile to update based on default shell
    DEFAULT_SHELL=$(basename "$SHELL")
    
    if [ "$DEFAULT_SHELL" = "zsh" ]; then
        PROFILE_FILE="$HOME/.zshrc"
    elif [ "$DEFAULT_SHELL" = "bash" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            PROFILE_FILE="$HOME/.bashrc"
        else
            PROFILE_FILE="$HOME/.bash_profile"
        fi
    else
        PROFILE_FILE="$HOME/.profile"
    fi
    
    echo "Detected shell: $DEFAULT_SHELL"
    echo "Will update: $PROFILE_FILE"
    
    # Add to profile file if not already present
    if ! grep -q "export PATH.*$DIST_PATH" "$PROFILE_FILE" 2>/dev/null; then
        echo "" >> "$PROFILE_FILE"
        echo "# Added by rfh install script" >> "$PROFILE_FILE"
        echo "export PATH=\"\$PATH:$DIST_PATH\"" >> "$PROFILE_FILE"
        echo "Successfully added $DIST_PATH to $PROFILE_FILE"
    else
        echo "PATH entry already exists in $PROFILE_FILE"
    fi
    
    # Add to current session PATH
    export PATH="$PATH:$DIST_PATH"
    echo "Added to current session PATH"
    echo "You may need to restart your terminal or run 'source $PROFILE_FILE' for the changes to take effect"
fi

echo ""
echo "Installation complete!"
echo "You can now run 'rfh' from anywhere in your terminal"
echo ""