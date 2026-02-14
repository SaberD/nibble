#!/usr/bin/env python3
"""
Nibble wrapper for pip.
Requires python-go which auto-installs Go.
Uses 'go install' to install nibble from source.
"""

import subprocess
import sys

def install_nibble():
    """Install nibble using go install"""
    print("Installing nibble using 'go install'...")
    result = subprocess.run(
        ["go", "install", "github.com/saberd/nibble@latest"],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        print(f"Error installing nibble: {result.stderr}", file=sys.stderr)
        sys.exit(1)
    
    print("✓ Installed successfully")

def main():
    """Main entry point"""
    install_nibble()

if __name__ == "__main__":
    main()
