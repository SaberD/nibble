import os
import platform
import shutil
import tarfile
import tempfile
from importlib import metadata
from pathlib import Path

from YOUR_PROJECT_cli.checksum import verify_release_checksum
from YOUR_PROJECT_cli.download import download_asset, download_text

OWNER = "YOUR_OWNER"
PROJECT = "YOUR_PROJECT"
DIST_NAME = "YOUR_PROJECT-cli"


def platform_triplet():
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


def install_dir():
    if os.name == "nt":
        base = Path(os.environ.get("LOCALAPPDATA", str(Path.home())))
    else:
        base = Path.home() / ".local" / "share"
    path = base / PROJECT
    path.mkdir(parents=True, exist_ok=True)
    return path


def dist_version():
    version = metadata.version(DIST_NAME)
    return version[1:] if version.startswith("v") else version


def binary_name():
    return f"{PROJECT}.exe" if os.name == "nt" else PROJECT


def release_base(version):
    return f"https://github.com/{OWNER}/{PROJECT}/releases/download/v{version}"


def extract_binary(archive_path, dest_binary):
    wanted = {binary_name(), PROJECT, f"{PROJECT}.exe"}

    with tarfile.open(archive_path, "r:*") as tf:
        member = next(
            (m for m in tf.getmembers() if m.isfile() and Path(m.name).name in wanted),
            None,
        )
        if member is None:
            raise RuntimeError("binary not found inside release archive")

        src = tf.extractfile(member)
        if src is None:
            raise RuntimeError("failed to extract binary from release archive")

        with src, open(dest_binary, "wb") as dst:
            shutil.copyfileobj(src, dst)


def ensure_installed():
    version = dist_version()
    os_name, arch = platform_triplet()

    # Keep binaries in a versioned local cache directory.
    install_root = install_dir()
    binary_path = install_root / version / "bin" / binary_name()
    binary_path.parent.mkdir(parents=True, exist_ok=True)
    if binary_path.exists():
        return binary_path

    asset = f"{PROJECT}_{os_name}_{arch}.tar.gz"
    release_url = release_base(version)
    archive_url = f"{release_url}/{asset}"

    with tempfile.TemporaryDirectory() as tmp:
        # Download and verify the release asset before extracting it.
        archive_path = Path(tmp) / asset
        if not download_asset(archive_url, archive_path):
            raise RuntimeError(
                f"release asset not found for {os_name}/{arch} at v{version}: {asset}"
            )
        verify_release_checksum(
            base_url=release_url,
            project=PROJECT,
            version=version,
            archive_name=asset,
            archive_path=archive_path,
            download_text=download_text,
        )
        # Extract the YOUR_PROJECT binary from the archive into the cache path.
        extract_binary(archive_path, binary_path)

    if os.name != "nt":
        binary_path.chmod(0o755)
    return binary_path
