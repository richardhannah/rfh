const { Before, After, setWorldConstructor } = require('@cucumber/cucumber');
const CustomWorld = require('./world');

setWorldConstructor(CustomWorld);

Before(function () {
  // Initialize any test state
  this.testState = {};
});

After(async function () {
  // Clean up after each scenario
  await this.cleanup();
});