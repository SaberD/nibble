#!/usr/bin/env python3
import argparse
import os
import shutil
import sys
from importlib import metadata
from pathlib import Path

PROJECT = "YOUR_PROJECT"
DIST_NAME = "YOUR_PROJECT-cli"


def cache_root():
    if os.name == "nt":
        base = Path(os.environ.get("LOCALAPPDATA", str(Path.home())))
    else:
        base = Path.home() / ".local" / "share"
    return base / PROJECT


def installed_version():
    try:
        version = metadata.version(DIST_NAME)
    except metadata.PackageNotFoundError:
        return None
    return version[1:] if version.startswith("v") else version


def remove_path(path):
    if not path.exists():
        return False
    shutil.rmtree(path)
    return True


def main():
    parser = argparse.ArgumentParser(description="Clean cached YOUR_PROJECT binaries.")
    parser.add_argument(
        "--all",
        action="store_true",
        help="Remove all cached YOUR_PROJECT versions.",
    )
    args = parser.parse_args()

    root = cache_root()
    if args.all:
        removed = remove_path(root)
        if removed:
            print(f"Removed cache: {root}")
        else:
            print(f"No cache found at: {root}")
        return 0

    version = installed_version()
    if version is None:
        print("YOUR_PROJECT-cli is not installed. Use --all to remove cache root.", file=sys.stderr)
        return 1

    target = root / version
    removed = remove_path(target)
    if removed:
        print(f"Removed cache for v{version}: {target}")
    else:
        print(f"No cache found for v{version}: {target}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
