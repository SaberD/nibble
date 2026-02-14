#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');
const os = require('os');

// The Go binary is installed to $GOPATH/bin or ~/go/bin
const goBin = path.join(os.homedir(), 'go', 'bin', 'nibble');

try {
  const args = process.argv.slice(2).join(' ');
  execSync(`${goBin} ${args}`, { stdio: 'inherit' });
} catch (error) {
  process.exit(error.status || 1);
}
