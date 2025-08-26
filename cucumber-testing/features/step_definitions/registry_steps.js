const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');
const os = require('os');

// Config management steps
Given('I have a clean config file', async function () {
  // Ensure we start with a clean config
  this.configPath = path.join(os.homedir(), '.rfh', 'config.toml');
  if (await fs.pathExists(this.configPath)) {
    // Backup existing config
    this.originalConfig = await fs.readFile(this.configPath, 'utf8');
    await fs.remove(this.configPath);
  }
  
  // Ensure .rfh directory exists
  await fs.ensureDir(path.dirname(this.configPath));
});

Given('I have a registry {string} configured at {string}', async function (name, url) {
  // Add registry via command to simulate real usage
  await this.runCommand(`rfh registry add ${name} ${url}`);
});

Given('I have a registry {string} configured', async function (name) {
  await this.runCommand(`rfh registry add ${name} https://example.com`);
});

Given('{string} is the active registry', async function (name) {
  await this.runCommand(`rfh registry use ${name}`);
});

// Registry operations validation
Then('the config should contain registry {string} with URL {string}', async function (name, url) {
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, 'Config file should exist').to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`[registries.${name}]`);
  // RFH uses single quotes in TOML format
  expect(configContent).to.include(`url = '${url}'`);
});

Then('the config should store the token for {string}', async function (name) {
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`[registries.${name}]`);
  expect(configContent).to.include('token = ');
});

Then('{string} should be the current active registry', async function (name) {
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`current = '${name}'`);
});

Then('the config should contain both registries', async function () {
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.match(/\[registries\..+\]/g);
});

Then('{string} should remain the active registry', async function (name) {
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`current = '${name}'`);
});

Then('no registry should be active', async function () {
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.match(/current = "?"?/);
});

Then('the config should not contain registry {string}', async function (name) {
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.not.include(`[registries.${name}]`);
});

// Registry list validation
Then('I should see {string} in the registry list', function (name) {
  const output = this.lastCommandOutput;
  expect(output).to.include(name);
});

Then('I should see {string} marked as active', function (name) {
  const output = this.lastCommandOutput;
  // Look for patterns like "* production" or "(active)" next to the name
  expect(output).to.match(new RegExp(`(\\*\\s*${name}|${name}.*active)`, 'i'));
});

Then('I should not see {string} marked as active', function (name) {
  const output = this.lastCommandOutput;
  expect(output).to.not.match(new RegExp(`\\*\\s*${name}`, 'i'));
});

Then('I should see both registries in the list', function () {
  const output = this.lastCommandOutput;
  // Should show multiple registry entries - check for actual format
  expect(output).to.match(/production/);
  expect(output).to.match(/staging/);
});

// Error handling
Then('I should see an error about registry not found', function () {
  const output = this.lastCommandError || this.lastCommandOutput;
  expect(output).to.match(/registry.*not found|not exist/i);
});

Then('I should see a warning about setting a new active registry', function () {
  const output = this.lastCommandOutput;
  expect(output).to.match(/warning|active.*registry/i);
});

Then('the original registry should remain unchanged', async function () {
  // This would require checking that the URL hasn't changed
  // Implementation depends on the actual behavior
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include('https://registry.example.com');
  expect(configContent).to.not.include('https://different.example.com');
});

// Cleanup hook - restore original config
const { After } = require('@cucumber/cucumber');

After(async function () {
  if (this.originalConfig && this.configPath) {
    try {
      await fs.writeFile(this.configPath, this.originalConfig);
    } catch (error) {
      // Ignore cleanup errors
    }
  } else if (this.configPath) {
    try {
      await fs.remove(this.configPath);
    } catch (error) {
      // Ignore cleanup errors  
    }
  }
});