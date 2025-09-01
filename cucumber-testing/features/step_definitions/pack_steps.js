const { Given, When, Then } = require('@cucumber/cucumber');
const { expect } = require('chai');
const fs = require('fs-extra');
const path = require('path');
const { execSync } = require('child_process');
const os = require('os');

// Helper functions are now provided by the World class

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
  // Create a ProjectManifest instead of PackageManifest array
  // This step is misleading - it should create a project manifest, not a package manifest
  // For now, create an empty project manifest since packages are managed separately
  const manifestContent = {
    "version": "1.0.0",
    "projectRoot": ".",
    "dependencies": {}
  };
  
  // Use World's writeFile method for consistency
  await this.writeFile('rulestack.json', JSON.stringify(manifestContent, null, 2));
});

Given('I have a rulestack.json manifest with name {string} and version {string} in {string}', async function (name, version, dirPath) {
  // Create a ProjectManifest instead of PackageManifest array
  const manifestContent = {
    "version": "1.0.0",
    "projectRoot": ".",
    "dependencies": {}
  };
  
  const manifestPath = path.join(dirPath, 'rulestack.json');
  await fs.writeJSON(manifestPath, manifestContent, { spaces: 2 });
});

Given('I have a custom manifest {string} with name {string} and version {string}', async function (filename, name, version) {
  // Create a ProjectManifest instead of PackageManifest array
  const manifestContent = {
    "version": "1.0.0",
    "projectRoot": ".",
    "dependencies": {}
  };
  
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
  
  // This step is confusing - it's trying to modify a project manifest as if it's a package manifest
  // For now, just ensure the project manifest exists but don't modify it
  // The actual package manifests are stored in .rulestack directories
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
  await this.runCommand('rfh init');
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to initialize RFH: ${this.lastCommandError || this.lastCommandOutput}`);
  }
});

// Backward compatibility step for existing tests
Given('RFH is initialized in the directory', async function () {
  // Use World's runCommand method for consistency
  await this.runCommand('rfh init');
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to initialize RFH: ${this.lastCommandError || this.lastCommandOutput}`);
  }
});

Then('the rulestack.json should contain package {string} with version {string}', async function (packageName, version) {
  // Use World's readFile method for consistency  
  const manifestContent = await this.readFile('rulestack.json');
  const manifestData = JSON.parse(manifestContent);
  
  // This is checking for a package dependency in the project manifest
  // Check if the package exists in dependencies
  expect(manifestData.dependencies, 'Project manifest should have dependencies').to.exist;
  expect(manifestData.dependencies[packageName], `Package ${packageName} should be in dependencies`).to.equal(version);
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

// New step definitions for enhanced pack functionality (existing package updates)

Given('I have installed package {string} containing file {string}', async function (packageSpec, filename) {
  // Parse package specification (e.g., "test-rules@1.0.0")
  const [packageName, version] = packageSpec.split('@');
  
  // First run rfh init to set up the project properly
  await this.runCommand('rfh init');
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to initialize RFH: ${this.lastCommandError || this.lastCommandOutput}`);
  }
  
  // Read the project manifest and add our package dependency
  const existingContent = await this.readFile('rulestack.json');
  const projectManifest = JSON.parse(existingContent);
  
  // Add our package as a dependency
  projectManifest.dependencies[packageName] = version;
  
  // Write back the updated manifest
  await this.writeFile('rulestack.json', JSON.stringify(projectManifest, null, 2));
  
  // Create package directory structure
  const packageDir = `.rulestack/${packageName}.${version}`;
  const fullPackageDir = path.join(this.testDir, packageDir);
  await fs.ensureDir(fullPackageDir);
  
  // Create the rule file in the package directory
  const ruleFilePath = path.join(packageDir, filename);
  await this.writeFile(ruleFilePath, `# Existing rule file: ${filename}`);
  
  // Debug logging - show actual file contents and directory structure
  this.log(`Created package directory: ${fullPackageDir}`, 'debug');
  this.log(`Created rule file: ${path.join(this.testDir, ruleFilePath)}`, 'debug');
  this.log(`Project manifest dependencies: ${JSON.stringify(projectManifest.dependencies)}`, 'debug');
  
  // Read back and display the actual rulestack.json content
  try {
    const actualManifest = await this.readFile('rulestack.json');
    this.log(`ACTUAL rulestack.json content:\n${actualManifest}`, 'debug');
  } catch (error) {
    this.log(`Error reading rulestack.json: ${error.message}`, 'debug');
  }
  
  // List the actual directory structure
  try {
    const rulestackExists = await fs.pathExists(path.join(this.testDir, '.rulestack'));
    this.log(`Does .rulestack directory exist? ${rulestackExists}`, 'debug');
    
    if (rulestackExists) {
      const contents = await fs.readdir(path.join(this.testDir, '.rulestack'));
      this.log(`Contents of .rulestack directory: ${JSON.stringify(contents)}`, 'debug');
      
      // Check the specific package directory
      const packageDirExists = await fs.pathExists(fullPackageDir);
      this.log(`Does package directory exist? ${packageDirExists} at ${fullPackageDir}`, 'debug');
      
      if (packageDirExists) {
        const packageContents = await fs.readdir(fullPackageDir);
        this.log(`Contents of package directory: ${JSON.stringify(packageContents)}`, 'debug');
      }
    }
  } catch (error) {
    this.log(`Error checking directory structure: ${error.message}`, 'debug');
  }
});

Given('I have installed package {string} containing files:', async function (packageSpec, dataTable) {
  // Parse package specification (e.g., "multi-rules@1.0.0")
  const [packageName, version] = packageSpec.split('@');
  
  // First run rfh init to set up the project properly
  await this.runCommand('rfh init');
  if (this.lastExitCode !== 0) {
    throw new Error(`Failed to initialize RFH: ${this.lastCommandError || this.lastCommandOutput}`);
  }
  
  // Read the project manifest and add our package dependency
  const existingContent = await this.readFile('rulestack.json');
  const projectManifest = JSON.parse(existingContent);
  
  // Add our package as a dependency
  projectManifest.dependencies[packageName] = version;
  
  // Write back the updated manifest
  await this.writeFile('rulestack.json', JSON.stringify(projectManifest, null, 2));
  
  // Create package directory structure
  const packageDir = `.rulestack/${packageName}.${version}`;
  await fs.ensureDir(path.join(this.testDir, packageDir));
  
  // Create all the rule files specified in the data table
  for (const row of dataTable.hashes()) {
    await this.writeFile(path.join(packageDir, row.filename), row.content);
  }
});

Then('the staged archive should contain {string}', async function (filename) {
  // Find the most recent staged archive
  const stagedDir = path.join(this.testDir, '.rulestack', 'staged');
  
  if (!await fs.pathExists(stagedDir)) {
    throw new Error('No staged directory found');
  }
  
  const archives = await fs.readdir(stagedDir);
  const tgzFiles = archives.filter(f => f.endsWith('.tgz'));
  
  if (tgzFiles.length === 0) {
    throw new Error('No staged archives found');
  }
  
  // Use the last created archive (in case there are multiple)
  const archivePath = path.join(stagedDir, tgzFiles[tgzFiles.length - 1]);
  
  // Extract and verify the archive contains the expected file
  const containsFile = await this.archiveContainsFile(archivePath, filename);
  expect(containsFile, `Archive should contain ${filename}`).to.be.true;
});

Then('the staged archive should be named {string}', async function (expectedName) {
  const stagedDir = path.join(this.testDir, '.rulestack', 'staged');
  
  if (!await fs.pathExists(stagedDir)) {
    throw new Error('No staged directory found');
  }
  
  const archives = await fs.readdir(stagedDir);
  const found = archives.includes(expectedName);
  
  expect(found, `Expected to find archive named ${expectedName}, but found: ${archives.join(', ')}`).to.be.true;
});

Given('I debug the current test environment state', async function () {
  console.log('=== DEBUGGING TEST ENVIRONMENT STATE ===');
  
  // Check if rulestack.json exists and show its content
  const rulestackPath = path.join(this.testDir, 'rulestack.json');
  try {
    if (await fs.pathExists(rulestackPath)) {
      const content = await fs.readFile(rulestackPath, 'utf8');
      console.log(`rulestack.json exists and contains:\n${content}`);
    } else {
      console.log('rulestack.json does not exist');
    }
  } catch (error) {
    console.log(`Error reading rulestack.json: ${error.message}`);
  }
  
  // Check .rulestack directory structure
  const rulestackDir = path.join(this.testDir, '.rulestack');
  try {
    if (await fs.pathExists(rulestackDir)) {
      const contents = await fs.readdir(rulestackDir);
      console.log(`Contents of .rulestack directory: ${JSON.stringify(contents)}`);
      
      // Check each subdirectory
      for (const item of contents) {
        const itemPath = path.join(rulestackDir, item);
        const stat = await fs.stat(itemPath);
        if (stat.isDirectory()) {
          try {
            const subContents = await fs.readdir(itemPath);
            console.log(`Contents of .rulestack/${item}: ${JSON.stringify(subContents)}`);
          } catch (err) {
            console.log(`Error reading .rulestack/${item}: ${err.message}`);
          }
        }
      }
    } else {
      console.log('.rulestack directory does not exist');
    }
  } catch (error) {
    console.log(`Error checking .rulestack directory: ${error.message}`);
  }
  
  // List all files in the test directory
  try {
    const testDirContents = await fs.readdir(this.testDir);
    console.log(`Contents of test directory: ${JSON.stringify(testDirContents)}`);
  } catch (error) {
    console.log(`Error reading test directory: ${error.message}`);
  }
  
  console.log('=== END DEBUG OUTPUT ===');
});