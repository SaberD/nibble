from setuptools import setup
import os
import subprocess

def get_version():
    """Get version from git tags"""
    try:
        version = subprocess.check_output(
            ["git", "describe", "--tags", "--abbrev=0"],
            cwd=os.path.dirname(__file__),
            stderr=subprocess.DEVNULL
        ).decode().strip()
        return version if version.startswith("v") else version
    except:
        return "0.1.0"

setup(
    name="nibble-cli",
    version=get_version(),
    description="Fast local network scanner with hardware identification and a terminal UI",
    long_description=open("README.md").read() if os.path.exists("README.md") else "",
    long_description_content_type="text/markdown",
    author="saberd",
    author_email="your-email@example.com",
    url="https://github.com/saberd/nibble",
    py_modules=["nibble_installer"],
    entry_points={
        "console_scripts": [
            "nibble=nibble_installer:main",
        ],
    },
    install_requires=["python-go"],
    python_requires=">=3.6",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Environment :: Console",
        "Intended Audience :: System Administrators",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Topic :: System :: Networking",
    ],
)
