#!/usr/bin/env python3
"""
Nibble CLI wrapper for pip.
Downloads the matching GitHub Release binary and executes it.
"""

import os
import platform
import shutil
import subprocess
import sys
import tarfile
import tempfile
import urllib.error
import urllib.request
import zipfile
import hashlib
import re
from importlib import metadata
from pathlib import Path

REPO = "backendsystems/nibble"
PROJECT = "nibble"
DIST_NAME = "nibble-cli"


def _platform_triplet():
    system = platform.system().lower()
    machine = platform.machine().lower()

    os_map = {
        "linux": "linux",
        "darwin": "darwin",
        "windows": "windows",
    }
    arch_map = {
        "x86_64": "amd64",
        "amd64": "amd64",
        "aarch64": "arm64",
        "arm64": "arm64",
    }

    os_name = os_map.get(system)
    arch = arch_map.get(machine)
    if not os_name or not arch:
        raise RuntimeError(f"unsupported platform: system={system}, arch={machine}")
    return os_name, arch


def _install_dir():
    if os.name == "nt":
        base = Path(os.environ.get("LOCALAPPDATA", str(Path.home())))
    else:
        base = Path.home() / ".local" / "share"
    path = base / PROJECT
    path.mkdir(parents=True, exist_ok=True)
    return path


def _dist_version():
    version = metadata.version(DIST_NAME)
    return version[1:] if version.startswith("v") else version


def _binary_name():
    return f"{PROJECT}.exe" if os.name == "nt" else PROJECT


def _download_asset(url, out_path):
    try:
        with urllib.request.urlopen(url) as resp, open(out_path, "wb") as f:
            shutil.copyfileobj(resp, f)
        return True
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return False
        raise

def _download_text(url):
    try:
        with urllib.request.urlopen(url) as resp:
            return resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return None
        raise

def _parse_checksums(text):
    checksums = {}
    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        match = re.match(r"^([A-Fa-f0-9]{64})\s+\*?(.+)$", line)
        if not match:
            continue
        checksums[match.group(2).strip()] = match.group(1).lower()
    return checksums

def _sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 64), b""):
            h.update(chunk)
    return h.hexdigest()

def _verify_checksum(version, archive_name, archive_path):
    base_url = f"https://github.com/{REPO}/releases/download/v{version}"
    checksum_candidates = [
        "checksums.txt",
        f"{PROJECT}_{version}_checksums.txt",
    ]

    checksum_text = None
    checksum_name = None
    for candidate in checksum_candidates:
        text = _download_text(f"{base_url}/{candidate}")
        if text is not None:
            checksum_text = text
            checksum_name = candidate
            break

    if checksum_text is None:
        raise RuntimeError(
            f"no checksum file found for v{version} (tried: {', '.join(checksum_candidates)})"
        )

    checksums = _parse_checksums(checksum_text)
    expected = checksums.get(archive_name)
    if expected is None:
        raise RuntimeError(f"checksum for {archive_name} not found in {checksum_name}")

    actual = _sha256_file(archive_path)
    if actual != expected:
        raise RuntimeError(f"checksum mismatch for {archive_name}")


def _extract_binary(archive_path, dest_binary):
    bin_names = {_binary_name(), PROJECT, f"{PROJECT}.exe"}

    if str(archive_path).endswith(".zip"):
        with zipfile.ZipFile(archive_path, "r") as zf:
            for member in zf.namelist():
                name = Path(member).name
                if name in bin_names:
                    with zf.open(member) as src, open(dest_binary, "wb") as dst:
                        shutil.copyfileobj(src, dst)
                    return
    else:
        with tarfile.open(archive_path, "r:*") as tf:
            for member in tf.getmembers():
                if not member.isfile():
                    continue
                name = Path(member.name).name
                if name in bin_names:
                    src = tf.extractfile(member)
                    if src is None:
                        continue
                    with src, open(dest_binary, "wb") as dst:
                        shutil.copyfileobj(src, dst)
                    return

    raise RuntimeError("binary not found inside release archive")


def ensure_installed():
    install_dir = _install_dir()
    version = _dist_version()
    version_dir = install_dir / version / "bin"
    version_dir.mkdir(parents=True, exist_ok=True)
    binary_path = version_dir / _binary_name()
    if binary_path.exists():
        return binary_path

    os_name, arch = _platform_triplet()
    asset_base = f"{PROJECT}_{os_name}_{arch}"
    candidates = [f"{asset_base}.tar.gz", f"{asset_base}.zip"]

    with tempfile.TemporaryDirectory() as tmp:
        tmpdir = Path(tmp)
        archive_path = None
        for asset in candidates:
            url = f"https://github.com/{REPO}/releases/download/v{version}/{asset}"
            local = tmpdir / asset
            if _download_asset(url, local):
                archive_path = local
                _verify_checksum(version, asset, local)
                break
        if archive_path is None:
            raise RuntimeError(
                f"no release asset found for {os_name}/{arch} at v{version} (tried: {', '.join(candidates)})"
            )

        _extract_binary(archive_path, binary_path)

    if os.name != "nt":
        binary_path.chmod(0o755)
    return binary_path


def main():
    try:
        binary = ensure_installed()
    except Exception as e:
        print(f"nibble install error: {e}", file=sys.stderr)
        return 1

    result = subprocess.run([str(binary)] + sys.argv[1:])
    return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
