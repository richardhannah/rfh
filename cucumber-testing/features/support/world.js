const { World } = require('@cucumber/cucumber');
const fs = require('fs-extra');
const path = require('path');
const { execSync } = require('child_process');
const os = require('os');

class CustomWorld extends World {
  constructor(options) {
    super(options);
    this.testDir = null;
    this.originalDir = process.cwd();
    this.lastCommandOutput = '';
    this.lastCommandError = '';
    this.lastExitCode = 0;
    // OS-specific binary name: rfh.exe on Windows, rfh on Unix/Mac
    const binaryName = process.platform === 'win32' ? 'rfh.exe' : 'rfh';
    this.rfhBinary = path.resolve(__dirname, '../../../dist', binaryName);
  }

  async createTempDirectory() {
    this.testDir = await fs.mkdtemp(path.join(os.tmpdir(), 'rfh-test-'));
    process.chdir(this.testDir);
  }

  async cleanup() {
    if (this.testDir) {
      process.chdir(this.originalDir);
      await fs.remove(this.testDir);
      this.testDir = null;
    }
  }

  async runCommand(command) {
    try {
      // Replace 'rfh' with the actual binary path (handle both 'rfh ' and 'rfh' at end of string)
      const actualCommand = command.replace(/^rfh(\s|$)/, `"${this.rfhBinary}"$1`);
      
      this.lastCommandOutput = execSync(actualCommand, {
        cwd: this.testDir || process.cwd(),
        encoding: 'utf8',
        timeout: 30000
      });
      this.lastExitCode = 0;
    } catch (error) {
      this.lastCommandError = error.stderr || error.message || '';
      this.lastCommandOutput = error.stdout || '';
      this.lastExitCode = error.status || 1;
    }
  }

  async fileExists(filePath) {
    const fullPath = this.testDir ? path.join(this.testDir, filePath) : filePath;
    return await fs.pathExists(fullPath);
  }

  async readFile(filePath) {
    const fullPath = this.testDir ? path.join(this.testDir, filePath) : filePath;
    return await fs.readFile(fullPath, 'utf8');
  }

  async writeFile(filePath, content) {
    const fullPath = this.testDir ? path.join(this.testDir, filePath) : filePath;
    await fs.writeFile(fullPath, content);
  }

  async directoryExists(dirPath) {
    const fullPath = this.testDir ? path.join(this.testDir, dirPath) : dirPath;
    const stat = await fs.stat(fullPath).catch(() => null);
    return stat && stat.isDirectory();
  }
}

module.exports = CustomWorld;