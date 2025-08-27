const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');
const { spawn } = require('child_process');

// Import shared helper functions
const { runCommand, runCommandInDirectory } = require('./helpers');

// Auth-specific step definitions for user registration testing

// Basic test for command availability
When('I register with username {string}, email {string}, and password {string}', async function (username, email, password) {
  // For scenarios testing basic registry/auth validation, use the simpler runCommand approach
  const command = `rfh auth register --username "${username}" --email "${email}" --password "${password}"`;
  await this.runCommand(command);
});

When('I register with username {string}, email {string}, password {string}, and confirmation {string}', async function (username, email, password, confirmation) {
  await this.runAuthRegisterTest();
});

When('I register with empty username, email {string}, and password {string}', async function (email, password) {
  await this.runAuthRegisterTest();
});

When('I register with username {string}, empty email, and password {string}', async function (username, password) {
  await this.runAuthRegisterTest();
});

// Login-specific step definitions
When('I login with username {string} and password {string}', async function (username, password) {
  // For scenarios testing basic registry/auth validation, use the simpler runCommand approach
  const command = `rfh auth login --username "${username}" --password "${password}"`;
  await this.runCommand(command);
});

// Register with credentials step definitions  
When('I register with username {string}, email {string}, and password {string} using flags', async function (username, email, password) {
  await this.runAuthRegisterWithCredentials(username, email, password);
});

// Interactive attempt step definitions
When('I attempt to login interactively', async function () {
  await this.runAuthLoginTest();
});

When('I attempt to register interactively', async function () {
  await this.runAuthRegisterTest();
});

Then('the config should contain the user {string} for registry {string}', async function (username, registryName) {
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, 'Config file should exist').to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`[registries.${registryName}]`);
  expect(configContent).to.include(`username = '${username}'`);
});

Then('the config should contain a JWT token for registry {string}', async function (registryName) {
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, 'Config file should exist').to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include(`[registries.${registryName}]`);
  expect(configContent).to.include('jwt_token = ');
});

Then('the config should contain global user {string}', async function (username) {
  const configExists = await fs.pathExists(this.configPath);
  expect(configExists, 'Config file should exist').to.be.true;
  
  const configContent = await fs.readFile(this.configPath, 'utf8');
  expect(configContent).to.include('[user]');
  expect(configContent).to.include(`username = '${username}'`);
  expect(configContent).to.include('token = ');
});

Then('the config should not contain any authentication data', async function () {
  const configExists = await fs.pathExists(this.configPath);
  if (configExists) {
    const configContent = await fs.readFile(this.configPath, 'utf8');
    expect(configContent).to.not.include('jwt_token = ');
    expect(configContent).to.not.include('username = ');
    expect(configContent).to.not.include('[user]');
  }
});

Then('the command should exit with zero status', function () {
  expect(this.lastCommandExitCode).to.equal(0);
});

Given('I have a clean config file with no registries', async function () {
  // Set a config path that doesn't exist - this will cause RFH to behave as if no config is present
  this.configPath = path.join(require('os').tmpdir(), `rfh-test-no-config-${Date.now()}-${Math.random().toString(36)}`, 'nonexistent.toml');
  
  // Don't create the config file - let RFH handle the absence of config
});

Given('I have a config with current registry {string}', async function (registryName) {
  const configDir = path.dirname(this.configPath);
  await fs.ensureDir(configDir);
  
  const configContent = `current = '${registryName}'

[registries]
`;
  await fs.writeFile(this.configPath, configContent, 'utf8');
});

// Simple test helper that focuses on error cases that don't require interactive input
async function runAuthRegisterTest() {
  const rfhPath = path.resolve(__dirname, '../../../dist/rfh');
  const configPath = this.configPath;
  
  return new Promise((resolve) => {
    const child = spawn(rfhPath, ['auth', 'register'], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(configPath)
    });
    
    let stdout = '';
    let stderr = '';
    
    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    
    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    // Immediately close stdin to avoid hanging on input prompts
    child.stdin.end();
    
    // Set a timeout to prevent hanging
    const timeout = setTimeout(() => {
      child.kill('SIGTERM');
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = child.exitCode || 1;
      resolve();
    }, 2000);
    
    child.on('close', (code) => {
      clearTimeout(timeout);
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = code;
      resolve();
    });
    
    child.on('error', (error) => {
      clearTimeout(timeout);
      this.lastCommandOutput = error.message;
      this.lastCommandExitCode = 1;
      resolve();
    });
  });
}

// Simple test helper for auth login command 
async function runAuthLoginTest() {
  const rfhPath = path.resolve(__dirname, '../../../dist/rfh');
  const configPath = this.configPath;
  
  return new Promise((resolve) => {
    const child = spawn(rfhPath, ['auth', 'login'], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(configPath)
    });
    
    let stdout = '';
    let stderr = '';
    
    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    
    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    // Immediately close stdin to avoid hanging on input prompts
    child.stdin.end();
    
    // Set a timeout to prevent hanging
    const timeout = setTimeout(() => {
      child.kill('SIGTERM');
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = child.exitCode || 1;
      resolve();
    }, 2000);
    
    child.on('close', (code) => {
      clearTimeout(timeout);
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = code;
      resolve();
    });
    
    child.on('error', (error) => {
      clearTimeout(timeout);
      this.lastCommandOutput = error.message;
      this.lastCommandExitCode = 1;
      resolve();
    });
  });
}

// Helper function for non-interactive auth login with credentials
async function runAuthLoginWithCredentials(username, password) {
  const rfhPath = path.resolve(__dirname, '../../../dist/rfh');
  const configPath = this.configPath;
  
  return new Promise((resolve) => {
    const child = spawn(rfhPath, [
      'auth', 'login',
      '--username', username,
      '--password', password
    ], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(configPath)
    });
    
    let stdout = '';
    let stderr = '';
    
    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    
    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    // Set a timeout 
    const timeout = setTimeout(() => {
      child.kill('SIGTERM');
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = child.exitCode || 1;
      resolve();
    }, 5000);
    
    child.on('close', (code) => {
      clearTimeout(timeout);
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = code;
      resolve();
    });
    
    child.on('error', (error) => {
      clearTimeout(timeout);
      this.lastCommandOutput = error.message;
      this.lastCommandExitCode = 1;
      resolve();
    });
  });
}

// Helper function for non-interactive auth register with credentials
async function runAuthRegisterWithCredentials(username, email, password) {
  const rfhPath = path.resolve(__dirname, '../../../dist/rfh');
  const configPath = this.configPath;
  
  return new Promise((resolve) => {
    const child = spawn(rfhPath, [
      'auth', 'register',
      '--username', username,
      '--email', email,
      '--password', password
    ], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(configPath)
    });
    
    let stdout = '';
    let stderr = '';
    
    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    
    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    // Set a timeout
    const timeout = setTimeout(() => {
      child.kill('SIGTERM');
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = child.exitCode || 1;
      resolve();
    }, 5000);
    
    child.on('close', (code) => {
      clearTimeout(timeout);
      this.lastCommandOutput = stdout + stderr;
      this.lastCommandExitCode = code;
      resolve();
    });
    
    child.on('error', (error) => {
      clearTimeout(timeout);
      this.lastCommandOutput = error.message;
      this.lastCommandExitCode = 1;
      resolve();
    });
  });
}

// Attach ALL helper functions to the world context (both shared and auth-specific)
require('@cucumber/cucumber').setDefinitionFunctionWrapper(function(fn) {
  return function(...args) {
    // Shared command execution functions
    this.runCommand = runCommand.bind(this);
    this.runCommandInDirectory = runCommandInDirectory.bind(this);
    
    // Auth-specific functions
    this.runAuthRegisterTest = runAuthRegisterTest.bind(this);
    this.runAuthLoginTest = runAuthLoginTest.bind(this);
    this.runAuthLoginWithCredentials = runAuthLoginWithCredentials.bind(this);
    this.runAuthRegisterWithCredentials = runAuthRegisterWithCredentials.bind(this);
    return fn.apply(this, args);
  };
});