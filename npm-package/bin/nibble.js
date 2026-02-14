#!/usr/bin/env node

const { spawnSync } = require('child_process');
const path = require('path');

const binName = process.platform === 'win32' ? 'nibble.exe' : 'nibble';
const binary = path.join(__dirname, '..', 'vendor', binName);

const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' });
if (result.error) {
  console.error(`Failed to execute nibble binary: ${result.error.message}`);
  process.exit(1);
}
process.exit(result.status || 0);
