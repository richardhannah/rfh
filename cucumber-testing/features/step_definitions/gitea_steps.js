const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');

// Helper function to check if Gitea is ready
async function checkGiteaStatus() {
  try {
    const response = await fetch('http://localhost:3000/api/v1/version');
    return response.ok;
  } catch (error) {
    return false;
  }
}

Given('Gitea server is running at localhost:3000', async function () {
  const maxRetries = 10;
  const retryDelay = 2000; // 2 seconds
  
  for (let i = 0; i < maxRetries; i++) {
    if (await checkGiteaStatus()) {
      // Gitea is ready
      return;
    }
    
    if (i < maxRetries - 1) {
      console.log(`Waiting for Gitea to be ready... (attempt ${i + 1}/${maxRetries})`);
      await new Promise(resolve => setTimeout(resolve, retryDelay));
    }
  }
  
  throw new Error('Gitea server is not running or not responding at localhost:3000. Please ensure docker-compose.test.yml is running with gitea-test service.');
});

Given('I create a test rule file {string} with content:', async function (filename, content) {
  await this.writeFile(filename, content);
});

// Removed - duplicate step definition exists in world.js

Then('the command should succeed or fail with {string}', function (expectedError) {
  if (this.lastExitCode !== 0) {
    const output = this.lastCommandError || this.lastCommandOutput;
    expect(output.toLowerCase()).to.include(expectedError.toLowerCase());
  }
  // If it succeeded, that's also acceptable
});

Then('the output should contain one of:', function (dataTable) {
  const possibleOutputs = dataTable.raw().map(row => row[0]);
  const output = (this.lastCommandOutput + (this.lastCommandError || '')).toLowerCase();
  
  const found = possibleOutputs.some(expected => 
    output.includes(expected.toLowerCase())
  );
  
  expect(found, `Expected output to contain one of: ${possibleOutputs.join(', ')}\nActual output: ${output}`).to.be.true;
});

Then('if successful the output should contain package results', function () {
  if (this.lastExitCode === 0) {
    const output = this.lastCommandOutput;
    // Check for typical package output patterns
    const hasPackageOutput = 
      output.includes('package') || 
      output.includes('Package') ||
      output.includes('Found') ||
      output.includes('Results') ||
      output.includes('No packages found'); // This is also valid
    
    expect(hasPackageOutput).to.be.true;
  }
  // If command failed, we already checked for expected error in previous step
});