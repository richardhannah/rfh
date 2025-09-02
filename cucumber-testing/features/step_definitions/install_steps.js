const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');

// Step definitions for install command tests

Given('I have a project manifest with the following dependencies:', async function (dataTable) {
  // Ensure we have a test directory and rulestack.json exists
  if (!this.testDir) {
    throw new Error('Test directory not initialized');
  }
  
  const manifestPath = path.join(this.testDir, 'rulestack.json');
  
  // Create dependencies object from table
  const dependencies = {};
  for (const row of dataTable.hashes()) {
    dependencies[row['package-name']] = row.version;
  }
  
  // Create project manifest
  const manifest = {
    version: "1.0.0",
    dependencies: dependencies
  };
  
  await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2));
});

Given('I have a project manifest with no dependencies', async function () {
  // Ensure we have a test directory and rulestack.json exists
  if (!this.testDir) {
    throw new Error('Test directory not initialized');
  }
  
  const manifestPath = path.join(this.testDir, 'rulestack.json');
  
  // Create project manifest with empty dependencies
  const manifest = {
    version: "1.0.0",
    dependencies: {}
  };
  
  await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2));
});

Given('I have already installed {string} version {string}', async function (packageName, version) {
  // Create .rulestack directory structure to simulate installed package
  const rulestackDir = path.join(this.testDir, '.rulestack');
  const packageDir = path.join(rulestackDir, `${packageName}.${version}`);
  
  await fs.ensureDir(packageDir);
  
  // Create a simple rule file to make the package look real
  const ruleFile = path.join(packageDir, `${packageName}_rules.md`);
  await fs.writeFile(ruleFile, `# ${packageName} Rules\n\nTest rules for ${packageName} v${version}`);
  
  // Update rulestack.json to include this dependency
  const manifestPath = path.join(this.testDir, 'rulestack.json');
  if (await fs.pathExists(manifestPath)) {
    const manifestContent = await fs.readFile(manifestPath, 'utf8');
    const manifest = JSON.parse(manifestContent);
    if (!manifest.dependencies) {
      manifest.dependencies = {};
    }
    manifest.dependencies[packageName] = version;
    await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2));
  }
  
  // Create or update lock file
  const lockPath = path.join(this.testDir, 'rulestack.lock.json');
  let lockManifest;
  
  if (await fs.pathExists(lockPath)) {
    const lockContent = await fs.readFile(lockPath, 'utf8');
    lockManifest = JSON.parse(lockContent);
  } else {
    lockManifest = {
      version: "1.0.0",
      packages: {}
    };
  }
  
  lockManifest.packages[packageName] = {
    version: version,
    sha256: "mock-sha256-hash"
  };
  
  await fs.writeFile(lockPath, JSON.stringify(lockManifest, null, 2));
});