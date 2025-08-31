const { Before, After, AfterAll, setWorldConstructor } = require('@cucumber/cucumber');
const CustomWorld = require('./world');
const fs = require('fs-extra');
const path = require('path');
const os = require('os');

setWorldConstructor(CustomWorld);

Before(async function () {
  // Initialize temp directory and test state for each scenario
  await this.createTempDirectory();
  this.testState = {};
});

After(async function () {
  // Clean up after each scenario
  await this.cleanup();
});

AfterAll(async function () {
  // Clean up shared cucumber config directory after all tests
  const cucumberConfigDir = path.join(os.homedir(), '.rfh-cucumber');
  try {
    await fs.remove(cucumberConfigDir);
    console.log('Cleaned up ~/.rfh-cucumber directory');
  } catch (error) {
    console.warn('Failed to cleanup ~/.rfh-cucumber:', error.message);
  }
});