#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const os = require('os');
const path = require('path');
const crypto = require('crypto');
const tar = require('tar');

const PROJECT = 'YOUR_PROJECT';
const OWNER = 'YOUR_OWNER';
const ROOT = path.resolve(__dirname, '..');
const VENDOR_DIR = path.join(ROOT, 'vendor');
const { version } = require(path.join(ROOT, 'package.json'));

const platformMap = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
const archMap = { x64: 'amd64', arm64: 'arm64' };

const osPlatform = platformMap[process.platform];
const osArch = archMap[process.arch];

if (!osPlatform || !osArch) {
  console.error(`Unsupported platform: ${process.platform}/${process.arch}`);
  process.exit(1);
}

const tag = `v${version}`;
const base = `https://github.com/${OWNER}/${PROJECT}/releases/download/${tag}`;
const archiveName = `${PROJECT}_${osPlatform}_${osArch}.tar.gz`;
const binName = process.platform === 'win32' ? `${PROJECT}.exe` : PROJECT;
const destPath = path.join(VENDOR_DIR, binName);

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    function get(url) {
      https.get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} downloading ${url}`));
        }
        res.pipe(file);
        file.on('finish', () => file.close(resolve));
        file.on('error', reject);
      }).on('error', reject);
    }
    get(url);
  });
}

function fetch(url) {
  return new Promise((resolve, reject) => {
    function get(url) {
      https.get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} fetching ${url}`));
        }
        const chunks = [];
        res.on('data', c => chunks.push(c));
        res.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
        res.on('error', reject);
      }).on('error', reject);
    }
    get(url);
  });
}

function sha256File(filePath) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash('sha256');
    fs.createReadStream(filePath)
      .on('data', d => hash.update(d))
      .on('end', () => resolve(hash.digest('hex')))
      .on('error', reject);
  });
}

function parseChecksum(checksumsTxt, filename) {
  for (const line of checksumsTxt.split('\n')) {
    const [hash, name] = line.trim().split(/\s+/);
    if (name === filename) return hash;
  }
  return null;
}

async function main() {
  console.log(`Installing ${PROJECT} ${version} for ${osPlatform}/${osArch}...`);
  fs.mkdirSync(VENDOR_DIR, { recursive: true });

  const checksumsTxt = await fetch(`${base}/${PROJECT}_${version}_checksums.txt`);
  const expectedHash = parseChecksum(checksumsTxt, archiveName);
  if (!expectedHash) {
    throw new Error(`No checksum found for ${archiveName} in checksums.txt`);
  }

  const tmpFile = path.join(os.tmpdir(), archiveName);
  try {
    await download(`${base}/${archiveName}`, tmpFile);

    const actualHash = await sha256File(tmpFile);
    if (actualHash !== expectedHash) {
      throw new Error(`Checksum mismatch for ${archiveName}: expected ${expectedHash}, got ${actualHash}`);
    }

    await tar.extract({ file: tmpFile, cwd: VENDOR_DIR, filter: p => p === binName });
    console.log(`Installed ${PROJECT} to ${destPath}`);
  } finally {
    try { fs.unlinkSync(tmpFile); } catch {}
  }
}

main().catch((err) => {
  console.error(`${PROJECT} install failed: ${err.message}`);
  process.exit(1);
});
