const { Given } = require('@cucumber/cucumber');
const fs = require('fs-extra');
const path = require('path');

// Step definitions for root user security warning tests
// Note: "I should see" and "I should not see" step definitions already exist in init_steps.js

Given('I am logged in as {string} user', async function (username) {
  // Ensure we have a test directory
  if (!this.testDir) {
    throw new Error('Test directory not initialized');
  }
  
  // Use the existing test config directory pattern from world.js
  const configDir = this.testConfigDir || path.join(this.testDir, '.rfh');
  await fs.ensureDir(configDir);
  
  // Create config file with the specified username
  const tomlContent = `current = "test-registry"

[registries]
[registries.test-registry]
url = "http://localhost:8081"
username = "${username}"
jwt_token = "mock-jwt-token"
`;
  
  await fs.writeFile(path.join(configDir, 'config.toml'), tomlContent);
  
  // Update testConfigDir to ensure commands use this config
  this.testConfigDir = configDir;
});

Given('I am not logged in to any registry', async function () {
  // Ensure we have a test directory
  if (!this.testDir) {
    throw new Error('Test directory not initialized');
  }
  
  // Create config directory but remove any config file
  const configDir = this.testConfigDir || path.join(this.testDir, '.rfh');
  await fs.ensureDir(configDir);
  await fs.remove(path.join(configDir, 'config.toml'));
  
  // Update testConfigDir
  this.testConfigDir = configDir;
});

Given('I have no active registry configured', async function () {
  // Ensure we have a test directory
  if (!this.testDir) {
    throw new Error('Test directory not initialized');
  }
  
  // Create config directory
  const configDir = this.testConfigDir || path.join(this.testDir, '.rfh');
  await fs.ensureDir(configDir);
  
  // Create config file with no current registry
  const tomlContent = `# No current registry configured

[registries]
# No registries configured
`;
  
  await fs.writeFile(path.join(configDir, 'config.toml'), tomlContent);
  
  // Update testConfigDir
  this.testConfigDir = configDir;
});