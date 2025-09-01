const { Before, After, AfterAll, setWorldConstructor } = require('@cucumber/cucumber');
const CustomWorld = require('./world');
const fs = require('fs-extra');
const path = require('path');
const os = require('os');

// Global delay configuration - adjust this value to control inter-scenario timing
// Can be overridden with environment variable: CUCUMBER_SCENARIO_DELAY=1000
const INTER_SCENARIO_DELAY_MS = parseInt(process.env.CUCUMBER_SCENARIO_DELAY) || 0; // 0 seconds default (no delay)

setWorldConstructor(CustomWorld);

Before(async function () {
  // Initialize temp directory and test state for each scenario
  await this.createTempDirectory();
  this.testState = {};
});

After(async function () {
  // Clean up after each scenario
  await this.cleanup();
  
  // Add delay between scenarios to prevent rate limiting on the test API
  if (INTER_SCENARIO_DELAY_MS > 0) {
    await new Promise(resolve => setTimeout(resolve, INTER_SCENARIO_DELAY_MS));
  }
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