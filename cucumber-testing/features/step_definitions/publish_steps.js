const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');
const os = require('os');

// Import shared helper functions
require('./helpers');

// Publish-specific step definitions

// "I have a clean config file with no registries" step is defined in auth_steps.js

Then('I should see either authentication error or connection error', function () {
  const output = this.lastCommandOutput;
  
  // Should see either an auth error or connection error since we're using dummy tokens/offline servers
  const hasExpectedError = output.includes('authentication') || 
                           output.includes('connection') ||
                           output.includes('unauthorized') ||
                           output.includes('forbidden') ||
                           output.includes('Invalid token') ||
                           output.includes('status 401') ||
                           output.includes('status 403') ||
                           output.includes('status 429') ||
                           output.includes('Rate limit') ||
                           output.includes('network') ||
                           output.includes('timeout') ||
                           output.includes('refused') ||
                           output.includes('unreachable') ||
                           output.includes('EOF');
                           
  expect(hasExpectedError, `Should see authentication or connection error, got: ${output}`).to.be.true;
});

// Cleanup hook for temporary directories
const { After } = require('@cucumber/cucumber');

After(async function () {
  // Clean up any temporary project directories created for publish tests
  if (this.tempProjectDir) {
    try {
      await fs.remove(this.tempProjectDir);
    } catch (error) {
      // Ignore cleanup errors
    }
  }
});