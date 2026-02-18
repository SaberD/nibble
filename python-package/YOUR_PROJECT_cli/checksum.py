import hashlib


def get_checksum(text, archive_name):
    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split()
        if len(parts) < 2:
            continue
        checksum = parts[0].lower()
        filename = parts[-1].lstrip("*")
        if filename == archive_name:
            return checksum
    return None


def sha256_file(path):
    with open(path, "rb") as f:
        return hashlib.sha256(f.read()).hexdigest()


def verify_release_checksum(base_url, project, version, archive_name, archive_path, download_text):
    checksum_candidates = [
        "checksums.txt",
        f"{project}_{version}_checksums.txt",
    ]

    checksum_text = None
    checksum_name = None
    for candidate in checksum_candidates:
        text = download_text(f"{base_url}/{candidate}")
        if text is not None:
            checksum_text = text
            checksum_name = candidate
            break

    if checksum_text is None:
        raise RuntimeError(
            f"no checksum file found for v{version} (tried: {', '.join(checksum_candidates)})"
        )

    expected = get_checksum(checksum_text, archive_name)
    if expected is None:
        raise RuntimeError(f"checksum for {archive_name} not found in {checksum_name}")

    actual = sha256_file(archive_path)
    if actual != expected:
        raise RuntimeError(f"checksum mismatch for {archive_name}")
