#!/usr/bin/env node

const fs = require('fs');
const os = require('os');
const path = require('path');
const https = require('https');
const { spawnSync } = require('child_process');

const PROJECT = 'nibble';
const OWNER = 'saberd';
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
  const url = `https://github.com/${OWNER}/${REPO}/releases/download/${tag}/${asset}`;

  const cacheDir = fs.mkdtempSync(path.join(os.tmpdir(), 'nibble-npm-'));
  const archivePath = path.join(cacheDir, asset);

  try {
    console.log(`Downloading ${asset} from ${url}`);
    await download(url, archivePath);
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
