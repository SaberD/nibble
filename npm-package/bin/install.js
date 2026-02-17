#!/usr/bin/env node

const fs = require('fs');
const os = require('os');
const path = require('path');
const https = require('https');
const crypto = require('crypto');
const { spawnSync } = require('child_process');

const PROJECT = 'nibble';
const OWNER = 'backendsystems';
const REPO = 'nibble';
const ROOT = path.resolve(__dirname, '..');
const VENDOR_DIR = path.join(ROOT, 'vendor');
const PACKAGE_JSON = require(path.join(ROOT, 'package.json'));

function mapPlatform() {
  const platformMap = {
    linux: 'linux',
    darwin: 'darwin',
    win32: 'windows',
  };

  const archMap = {
    x64: 'amd64',
    arm64: 'arm64',
  };

  const osName = platformMap[process.platform];
  const arch = archMap[process.arch];
  if (!osName || !arch) {
    throw new Error(`unsupported platform: ${process.platform}/${process.arch}`);
  }

  return { osName, arch };
}

function download(url, outFile) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        return resolve(download(res.headers.location, outFile));
      }

      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`download failed (${res.statusCode}): ${url}`));
      }

      const file = fs.createWriteStream(outFile);
      res.pipe(file);
      file.on('finish', () => file.close(() => resolve()));
      file.on('error', reject);
    });
    req.on('error', reject);
  });
}

function downloadText(url) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        return resolve(downloadText(res.headers.location));
      }

      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`download failed (${res.statusCode}): ${url}`));
      }

      let data = '';
      res.setEncoding('utf8');
      res.on('data', (chunk) => {
        data += chunk;
      });
      res.on('end', () => resolve(data));
      res.on('error', reject);
    });
    req.on('error', reject);
  });
}

function parseChecksums(text) {
  const checksums = new Map();
  for (const line of text.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) {
      continue;
    }

    const match = trimmed.match(/^([a-fA-F0-9]{64})\s+\*?(.+)$/);
    if (!match) {
      continue;
    }
    checksums.set(match[2].trim(), match[1].toLowerCase());
  }
  return checksums;
}

function sha256File(filePath) {
  const hash = crypto.createHash('sha256');
  hash.update(fs.readFileSync(filePath));
  return hash.digest('hex');
}

async function verifyArchiveChecksum(urlBase, archivePath, archiveName, version) {
  const checksumCandidates = [
    'checksums.txt',
    `${PROJECT}_${version}_checksums.txt`,
  ];

  let checksumFile = '';
  let checksumName = '';
  for (const candidate of checksumCandidates) {
    try {
      checksumFile = await downloadText(`${urlBase}/${candidate}`);
      checksumName = candidate;
      break;
    } catch (err) {
      if (!String(err.message).includes('download failed (404)')) {
        throw err;
      }
    }
  }

  if (!checksumFile) {
    throw new Error(`no checksum file found (tried: ${checksumCandidates.join(', ')})`);
  }

  const checksums = parseChecksums(checksumFile);
  const expected = checksums.get(archiveName);
  if (!expected) {
    throw new Error(`checksum for ${archiveName} not found in ${checksumName}`);
  }

  const actual = sha256File(archivePath);
  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${archiveName}`);
  }
}

function run(cmd, args) {
  const result = spawnSync(cmd, args, { stdio: 'inherit' });
  if (result.status !== 0) {
    throw new Error(`command failed: ${cmd} ${args.join(' ')}`);
  }
}

function installFromArchive(archivePath, osName) {
  fs.mkdirSync(VENDOR_DIR, { recursive: true });

  const binName = osName === 'windows' ? `${PROJECT}.exe` : PROJECT;
  const destBinary = path.join(VENDOR_DIR, binName);

  if (archivePath.endsWith('.zip')) {
    if (process.platform === 'win32') {
      run('powershell', [
        '-NoProfile',
        '-Command',
        `Expand-Archive -Path \"${archivePath}\" -DestinationPath \"${VENDOR_DIR}\" -Force`,
      ]);
    } else {
      run('unzip', ['-o', archivePath, '-d', VENDOR_DIR]);
    }
  } else {
    run('tar', ['-xzf', archivePath, '-C', VENDOR_DIR]);
  }

  if (!fs.existsSync(destBinary)) {
    throw new Error(`binary not found after extraction: ${destBinary}`);
  }

  if (process.platform !== 'win32') {
    fs.chmodSync(destBinary, 0o755);
  }
}

async function main() {
  const { osName, arch } = mapPlatform();
  const pkgVersion = PACKAGE_JSON.version.replace(/^v/, '');
  const tag = `v${pkgVersion}`;
  const assetBase = `${PROJECT}_${osName}_${arch}`;
  const ext = 'tar.gz';
  const asset = `${assetBase}.${ext}`;
  const urlBase = `https://github.com/${OWNER}/${REPO}/releases/download/${tag}`;
  const url = `${urlBase}/${asset}`;

  const cacheDir = fs.mkdtempSync(path.join(os.tmpdir(), 'nibble-npm-'));
  const archivePath = path.join(cacheDir, asset);

  try {
    console.log(`Downloading ${asset} from ${url}`);
    await download(url, archivePath);
    await verifyArchiveChecksum(urlBase, archivePath, asset, pkgVersion);
    installFromArchive(archivePath, osName);
    console.log('Installed nibble binary successfully');
  } finally {
    if (fs.rmSync) {
      fs.rmSync(cacheDir, { recursive: true, force: true });
    } else {
      fs.rmdirSync(cacheDir, { recursive: true });
    }
  }
}

main().catch((err) => {
  console.error(`nibble install failed: ${err.message}`);
  process.exit(1);
});
