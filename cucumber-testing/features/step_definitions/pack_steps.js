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
  // Use World's createTempDirectory method for consistency
  await this.createTempDirectory();
});

Given('I have a temporary project directory at {string}', async function (dirPath) {
  // Use World's createTempDirectory method, then optionally create subdirectory
  await this.createTempDirectory();
  if (dirPath !== this.testDir) {
    this.tempProjectDir = path.resolve(dirPath);
    await fs.ensureDir(this.tempProjectDir);
  }
});

Given('I have a rulestack.json manifest with name {string} and version {string}', async function (name, version) {
  const manifestContent = [{
    "name": name,
    "version": version,
    "description": `Test package ${name}`,
    "files": ["*.md"]
  }];
  
  // Use World's writeFile method for consistency
  await this.writeFile('rulestack.json', JSON.stringify(manifestContent, null, 2));
});

Given('I have a rulestack.json manifest with name {string} and version {string} in {string}', async function (name, version, dirPath) {
  const manifestContent = [{
    "name": name,
    "version": version,
    "description": `Test package ${name}`,
    "files": ["*.md"]
  }];
  
  const manifestPath = path.join(dirPath, 'rulestack.json');
  await fs.writeJSON(manifestPath, manifestContent, { spaces: 2 });
});

Given('I have a custom manifest {string} with name {string} and version {string}', async function (filename, name, version) {
  const manifestContent = [{
    "name": name,
    "version": version,
    "description": `Custom package ${name}`,
    "files": ["*.md"]
  }];
  
  const manifestPath = path.join(this.tempProjectDir, filename);
  await fs.writeJSON(manifestPath, manifestContent, { spaces: 2 });
});

Given('I have a rule file {string} with content {string}', async function (filename, content) {
  // Use World's writeFile method for consistency
  await this.writeFile(filename, content);
});

Given('I have a rule file {string} with content {string} in {string}', async function (filename, content, dirPath) {
  const filePath = path.join(dirPath, filename);
  await fs.writeFile(filePath, content, 'utf8');
});

Given('the manifest includes file {string}', async function (filename) {
  // Use World's readFile and writeFile methods for consistency
  const manifestContent = await this.readFile('rulestack.json');
  const manifest = JSON.parse(manifestContent);
  
  // Add specific file to files array instead of glob pattern
  manifest.files = [filename];
  
  await this.writeFile('rulestack.json', JSON.stringify(manifest, null, 2));
});

When('I run {string} in the project directory', async function (command) {
  // Use World's runCommand method which already uses testDir
  await this.runCommand(command);
});

When('I run {string} in the {string} directory', async function (command, dirName) {
  // If dirName is different from current test directory, create it
  if (dirName !== this.testDir) {
    const targetDir = path.resolve(dirName);
    await this.runCommand(command, { cwd: targetDir });
  } else {
    await this.runCommand(command);
  }
});

Then('the archive file {string} should exist', async function (filename) {
  // Use World's fileExists method for consistency
  const exists = await this.fileExists(filename);
  expect(exists, `Archive file ${filename} should exist`).to.be.true;
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

// New step definitions for enhanced pack functionality

Given('RFH is initialized in the directory for package creation', async function () {
  // Use World's runCommand method for consistency
  await this.runCommand('rfh init --package');
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to initialize RFH: ${this.lastCommandError || this.lastCommandOutput}`);
  }
});

// Backward compatibility step for existing tests
Given('RFH is initialized in the directory', async function () {
  // Use World's runCommand method for consistency
  await this.runCommand('rfh init --package');
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to initialize RFH: ${this.lastCommandError || this.lastCommandOutput}`);
  }
});

Then('the rulestack.json should contain package {string} with version {string}', async function (packageName, version) {
  // Use World's readFile method for consistency
  const manifestContent = await this.readFile('rulestack.json');
  const manifestData = JSON.parse(manifestContent);
  
  // Handle both array and object formats
  const manifests = Array.isArray(manifestData) ? manifestData : [manifestData];
  
  const foundPackage = manifests.find(m => m.name === packageName && m.version === version);
  expect(foundPackage, `Package ${packageName} v${version} not found in manifest`).to.exist;
});

Then('the archive file {string} should not exist', async function (archivePath) {
  // Use World's fileExists method for consistency
  const exists = await this.fileExists(archivePath);
  expect(exists, `Archive ${archivePath} should not exist but it does`).to.be.false;
});

Then('the directory {string} should not exist', async function (dirPath) {
  // Use World's directoryExists method for consistency
  const exists = await this.directoryExists(dirPath);
  expect(exists, `Directory ${dirPath} should not exist but it does`).to.be.false;
});

Then('the directory {string} should exist', async function (dirPath) {
  // Use World's directoryExists method for consistency
  const exists = await this.directoryExists(dirPath);
  expect(exists, `Directory ${dirPath} should exist but it doesn't`).to.be.true;
});

// Use the existing step definition but update it to use the new binary
// Remove this duplicate - the step already exists at line 81

// Helper functions are defined in auth_steps.js and bound via setDefinitionFunctionWrapper

// OS-agnostic error message checks
Then('I should see a file not found error', function () {
  const output = this.lastCommandOutput + this.lastCommandError;
  // Check for both Linux/Mac and Windows error messages
  const hasFileNotFoundError = 
    output.includes('no such file or directory') || 
    output.includes('The system cannot find the file specified') ||
    output.includes('cannot find the path specified');
  
  if (!hasFileNotFoundError) {
    const message = `
Expected a file not found error message.

ACTUAL OUTPUT (full):
----------------------------------------
${output}
----------------------------------------
`;
    throw new Error(message);
  }
});

// Cleanup hook for temporary directories and attach helper functions
const { After, Before } = require('@cucumber/cucumber');

// Cleanup is now handled by World's cleanup method in hooks.js

// Helper functions are bound via auth_steps.js setDefinitionFunctionWrapper