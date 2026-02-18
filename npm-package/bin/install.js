#!/usr/bin/env node

const path = require('path');
const BinWrapper = require('bin-wrapper');

const PROJECT = 'nibble';
const OWNER = 'backendsystems';
const ROOT = path.resolve(__dirname, '..');
const VENDOR_DIR = path.join(ROOT, 'vendor');
const PACKAGE_JSON = require(path.join(ROOT, 'package.json'));

function buildWrapper() {
  const tag = `v${PACKAGE_JSON.version}`;
  const base = `https://github.com/${OWNER}/${PROJECT}/releases/download/${tag}`;
  const binName = process.platform === 'win32' ? `${PROJECT}.exe` : PROJECT;
  const supportedOs = ['linux', 'darwin', 'win32'];
  const supportedArch = ['x64', 'arm64'];

  const wrapper = new BinWrapper({ skipCheck: true });
  for (const os of supportedOs) {
    let osTarget = os;
    if (os === 'win32') {
      osTarget = 'windows';
    }

    for (const arch of supportedArch) {
      let archTarget = arch;
      if (arch === 'x64') {
        archTarget = 'amd64';
      }

      wrapper.src(`${base}/${PROJECT}_${osTarget}_${archTarget}.tar.gz`, os, arch);
    }
  }

  return wrapper.dest(VENDOR_DIR).use(binName);
}

async function main() {
  console.log(`Installing ${PROJECT} ${PACKAGE_JSON.version}...`);
  const bin = buildWrapper();
  await bin.run();
  console.log('Installed nibble binary successfully');
}

main().catch((err) => {
  console.error(`nibble install failed: ${err.message}`);
  process.exit(1);
});
