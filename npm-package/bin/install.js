#!/usr/bin/env node

const { spawnSync } = require('child_process');

function installNibble() {
  console.log("Installing nibble using 'go install'...");
  const result = spawnSync('go', ['install', 'github.com/saberd/nibble@latest'], {
    stdio: 'inherit',
    shell: true
  });
  
  if (result.status !== 0) {
    console.error('Error installing nibble');
    process.exit(1);
  }
  
  console.log('✓ Installed successfully');
}

installNibble();

