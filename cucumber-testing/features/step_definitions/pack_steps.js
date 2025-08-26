const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');
const { execSync } = require('child_process');
const os = require('os');

// Import shared helper functions
require('./helpers');

// Pack-specific step definitions

Given('I have a temporary project directory', async function () {
  // Create a temporary directory for the project
  this.tempProjectDir = path.join(os.tmpdir(), `rfh-pack-test-${Date.now()}`);
  await fs.ensureDir(this.tempProjectDir);
});

Given('I have a temporary project directory at {string}', async function (dirPath) {
  // Create a relative path from current working directory 
  this.tempProjectDir = path.resolve(dirPath);
  await fs.ensureDir(this.tempProjectDir);
});

Given('I have a rulestack.json manifest with name {string} and version {string}', async function (name, version) {
  const manifestContent = {
    "name": name,
    "version": version,
    "description": `Test package ${name}`,
    "files": ["*.md"]
  };
  
  const manifestPath = path.join(this.tempProjectDir, 'rulestack.json');
  await fs.writeJSON(manifestPath, manifestContent, { spaces: 2 });
});

Given('I have a rulestack.json manifest with name {string} and version {string} in {string}', async function (name, version, dirPath) {
  const manifestContent = {
    "name": name,
    "version": version,
    "description": `Test package ${name}`,
    "files": ["*.md"]
  };
  
  const manifestPath = path.join(dirPath, 'rulestack.json');
  await fs.writeJSON(manifestPath, manifestContent, { spaces: 2 });
});

Given('I have a custom manifest {string} with name {string} and version {string}', async function (filename, name, version) {
  const manifestContent = {
    "name": name,
    "version": version,
    "description": `Custom package ${name}`,
    "files": ["*.md"]
  };
  
  const manifestPath = path.join(this.tempProjectDir, filename);
  await fs.writeJSON(manifestPath, manifestContent, { spaces: 2 });
});

Given('I have a rule file {string} with content {string}', async function (filename, content) {
  const filePath = path.join(this.tempProjectDir, filename);
  await fs.writeFile(filePath, content, 'utf8');
});

Given('I have a rule file {string} with content {string} in {string}', async function (filename, content, dirPath) {
  const filePath = path.join(dirPath, filename);
  await fs.writeFile(filePath, content, 'utf8');
});

Given('the manifest includes file {string}', async function (filename) {
  const manifestPath = path.join(this.tempProjectDir, 'rulestack.json');
  const manifest = await fs.readJSON(manifestPath);
  
  // Add specific file to files array instead of glob pattern
  manifest.files = [filename];
  
  await fs.writeJSON(manifestPath, manifest, { spaces: 2 });
});

When('I run {string} in the project directory', async function (command) {
  await this.runCommandInDirectory(command, this.tempProjectDir);
});

Then('the archive file {string} should exist', async function (filename) {
  // Check both in temp directory and as absolute path
  let archivePath = path.join(this.tempProjectDir, filename);
  let exists = await fs.pathExists(archivePath);
  
  // If not found in temp directory, check if it's an absolute path or relative to current working directory
  if (!exists) {
    archivePath = path.isAbsolute(filename) ? filename : path.resolve(filename);
    exists = await fs.pathExists(archivePath);
  }
  
  expect(exists, `Archive file ${filename} should exist at ${archivePath}`).to.be.true;
});

// File existence is handled by init_steps.js to avoid conflicts

Then('I should see an error about missing files', function () {
  const output = this.lastCommandOutput;
  // Check for common error patterns related to missing files
  const hasError = output.includes('no files matched') || 
                  output.includes('file not found') || 
                  output.includes('missing file') ||
                  output.includes('failed to pack files');
  expect(hasError, 'Should see an error about missing files').to.be.true;
});

// Helper functions are defined in auth_steps.js and bound via setDefinitionFunctionWrapper

// Cleanup hook for temporary directories and attach helper functions
const { After, Before } = require('@cucumber/cucumber');

After(async function () {
  // Clean up temporary project directory if it was created
  if (this.tempProjectDir) {
    try {
      await fs.remove(this.tempProjectDir);
    } catch (error) {
      // Ignore cleanup errors
      console.warn('Failed to cleanup temp directory:', error.message);
    }
  }
});

// Helper functions are bound via auth_steps.js setDefinitionFunctionWrapper