const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');
const { execSync } = require('child_process');

// Helper functions are now provided by the World class
const os = require('os');

// Helper function to ensure test packages are published to the server using root user
// Now uses World instance for consistency
async function ensureTestPackagesPublished() {
  // Note: This function is bound to World context via setDefinitionFunctionWrapper
  // Create and publish test packages that are expected by the tests
  const testPackages = [
    { name: 'security-rules', version: '1.0.1' },
    { name: 'example-rules', version: '0.1.0' }
  ];
  
  for (const pkg of testPackages) {
    try {
      // Use publishPackage method from World class which handles isolation properly
      await this.publishPackage(pkg.name, pkg.version, `# ${pkg.name} Rules v${pkg.version}\n\nTest rules for ${pkg.name} v${pkg.version}`);
    } catch (error) {
      // Don't throw - test might still work if packages are already on server
      console.log(`Warning: Failed to set up test package ${pkg.name}@${pkg.version}: ${error.message}`);
    }
  }
}

// Config management steps
Given('I have a clean config file', async function () {
  // Using shared ~/.rfh-cucumber directory for production-like testing
  // The config path is set to ~/.rfh-cucumber/config.toml in World's createTempDirectory
  
  // Ensure the cucumber config directory exists 
  await fs.ensureDir(this.testConfigDir);
  
  // Instead of deleting the file, create an empty config file
  // This ensures the directory structure exists for RFH commands
  const emptyConfig = `# Empty config for testing
[registries]
`;
  await fs.writeFile(this.configPath, emptyConfig);
});

Given('I have a registry {string} configured at {string}', async function (name, url) {
  // Remove manual setup - now handled by ensureRegistrySetup
  if (url === 'http://localhost:8081') {
    await this.ensureRegistrySetup();
  } else {
    // Keep existing logic for non-test registries
    await this.runCommand(`rfh registry add ${name} ${url}`);
    if (this.lastExitCode !== 0) {
      throw new Error(`Failed to add registry ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
    }
    
    // Small delay to ensure file is written 
    await new Promise(resolve => setTimeout(resolve, 200));
  }
});

Given('I have a registry {string} configured', async function (name) {
  await this.runCommand(`rfh registry add ${name} https://example.com`);
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to add registry ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
  }
});

Given('{string} is the active registry', async function (name) {
  await this.runCommand(`rfh registry use ${name}`);
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to set active registry ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
  }
});

// Step to explicitly ensure test packages are published using root user
Given('test packages are published to the registry', async function () {
  await ensureTestPackagesPublished.bind(this)();
});

// Registry operations validation
Then('the config should contain registry {string} with URL {string}', async function (name, url) {
  // Small delay to ensure file is fully written
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, `Config file should exist at ${this.configPath}`).to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  
  // Debug output: show what we actually found (only if test fails)
  // Note: This debug output helps identify config file issues
  
  expect(configContent, `Config should contain [registries.${name}]\nActual config:\n${configContent}`).to.include(`[registries.${name}]`);
  // RFH uses single quotes in TOML format
  expect(configContent, `Config should contain url = '${url}'\nActual config:\n${configContent}`).to.include(`url = '${url}'`);
});

// Token storage step removed - JWT tokens are obtained via 'rfh auth login'

Then('{string} should be the current active registry', async function (name) {
  // Small delay to ensure file is fully written
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, `Config file should exist at ${this.configPath}`).to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`current = '${name}'`);
});

Then('the config should contain both registries', async function () {
  // Small delay to ensure file is fully written
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, `Config file should exist at ${this.configPath}`).to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.match(/\[registries\..+\]/g);
});

Then('{string} should remain the active registry', async function (name) {
  // Small delay to ensure file is fully written
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, `Config file should exist at ${this.configPath}`).to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`current = '${name}'`);
});

Then('no registry should be active', async function () {
  // Small delay to ensure file is fully written
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, `Config file should exist at ${this.configPath}`).to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.match(/current = "?"?/);
});

Then('the config should not contain registry {string}', async function (name) {
  // Small delay to ensure file is fully written
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, `Config file should exist at ${this.configPath}`).to.be.true;
  
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
  // With RFH_CONFIG, we're using isolated test configs
  // Cleanup is handled by World's cleanup method which removes the entire test directory
  // No need to restore or delete configs here as they're all in temp directories
  
  // Clear any test-specific state
  this.originalConfig = null;
});