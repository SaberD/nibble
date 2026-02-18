#!/usr/bin/env node

const path = require('path');
const BinWrapper = require('bin-wrapper');

const PROJECT = 'nibble';
const OWNER = 'backendsystems';
const ROOT = path.resolve(__dirname, '..');
const VENDOR_DIR = path.join(ROOT, 'vendor');
const PACKAGE_JSON = require(path.join(ROOT, 'package.json'));

function buildWrapper() {
  const version = PACKAGE_JSON.version.replace(/^v/, '');
  const tag = `v${version}`;
  const base = `https://github.com/${OWNER}/${PROJECT}/releases/download/${tag}`;
  const binName = process.platform === 'win32' ? `${PROJECT}.exe` : PROJECT;

  return new BinWrapper({ skipCheck: true })
    .src(`${base}/${PROJECT}_linux_amd64.tar.gz`, 'linux', 'x64')
    .src(`${base}/${PROJECT}_linux_arm64.tar.gz`, 'linux', 'arm64')
    .src(`${base}/${PROJECT}_darwin_amd64.tar.gz`, 'darwin', 'x64')
    .src(`${base}/${PROJECT}_darwin_arm64.tar.gz`, 'darwin', 'arm64')
    .src(`${base}/${PROJECT}_windows_amd64.tar.gz`, 'win32', 'x64')
    .src(`${base}/${PROJECT}_windows_arm64.tar.gz`, 'win32', 'arm64')
    .dest(VENDOR_DIR)
    .use(binName);
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
