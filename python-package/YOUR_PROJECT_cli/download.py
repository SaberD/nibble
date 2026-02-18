import shutil
import urllib.error
import urllib.request


def download_asset(url, out_path):
    try:
        with urllib.request.urlopen(url) as resp, open(out_path, "wb") as f:
            shutil.copyfileobj(resp, f)
        return True
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return False
        raise


def download_text(url):
    try:
        with urllib.request.urlopen(url) as resp:
            return resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return None
        raise
