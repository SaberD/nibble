#!/usr/bin/env python3
import subprocess
import sys

from YOUR_PROJECT_cli.install import ensure_installed


def main():
    try:
        binary = ensure_installed()
    except Exception as e:
        print(f"YOUR_PROJECT install error: {e}", file=sys.stderr)
        return 1

    result = subprocess.run([str(binary)] + sys.argv[1:])
    return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
