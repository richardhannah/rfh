const { execSync } = require('child_process');
const path = require('path');

// Shared helper functions for all step definitions
async function runCommand(command) {
  // Use absolute path to rfh.exe based on the original project directory
  const rfhPath = path.resolve(__dirname, '../../../dist/rfh.exe');
  
  // Replace 'rfh' with full path to executable
  const fullCommand = command.replace(/^rfh\s/, `"${rfhPath}" `);
  
  try {
    const output = execSync(fullCommand, {
      encoding: 'utf8',
      stdio: 'pipe'
    });
    
    this.lastCommandOutput = output;
    this.lastCommandExitCode = 0;
    this.lastExitCode = 0;
  } catch (error) {
    this.lastCommandOutput = error.stdout + error.stderr;
    this.lastCommandExitCode = error.status || 1;
    this.lastExitCode = error.status || 1;
  }
}

async function runCommandInDirectory(command, directory) {
  // Use absolute path to rfh.exe based on the original project directory
  const rfhPath = path.resolve(__dirname, '../../../dist/rfh.exe');
  
  // Replace 'rfh' with full path to executable
  const fullCommand = command.replace(/^rfh\s/, `"${rfhPath}" `);
  
  try {
    const output = execSync(fullCommand, {
      cwd: directory,
      encoding: 'utf8',
      stdio: 'pipe'
    });
    
    this.lastCommandOutput = output;
    this.lastCommandExitCode = 0;
    this.lastExitCode = 0;
  } catch (error) {
    this.lastCommandOutput = error.stdout + error.stderr;
    this.lastCommandExitCode = error.status || 1;
    this.lastExitCode = error.status || 1;
  }
}

// Don't bind functions here - let auth_steps.js handle all bindings

module.exports = {
  runCommand,
  runCommandInDirectory
};