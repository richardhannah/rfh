const { World } = require('@cucumber/cucumber');
const fs = require('fs-extra');
const path = require('path');
const { execSync } = require('child_process');
const os = require('os');
const toml = require('toml');
// Use dynamic import for node-fetch in Node.js versions that support it
let fetch;
try {
  const nodeFetch = require('node-fetch');
  fetch = nodeFetch.default || nodeFetch;
} catch (err) {
  // Fallback for environments without node-fetch
  console.warn('node-fetch not available, API calls will not work');
}

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
    
    // Configuration paths
    this.configPath = path.join(os.homedir(), '.rfh', 'config.toml');
    
    // Enhanced World properties for authentication and registry management
    this.rootJwtToken = null;
    this.currentUser = null;
    this.registryConfigured = false;
    this.testUsers = new Map(); // Track created test users
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

  async runCommand(command, options = {}) {
    try {
      // Replace 'rfh' with the actual binary path (handle both 'rfh ' and 'rfh' at end of string)
      const actualCommand = command.replace(/^rfh(\s|$)/, `"${this.rfhBinary}"$1`);
      
      this.lastCommandOutput = execSync(actualCommand, {
        cwd: options.cwd || this.testDir || process.cwd(),
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

  // Enhanced World methods for authentication and registry management
  
  // Auto-setup registry and root authentication
  async ensureRegistrySetup() {
    if (!this.registryConfigured) {
      await this.runCommand('rfh registry add test-registry http://localhost:8081');
      await this.runCommand('rfh registry use test-registry');
      await this.loginAsRoot();
      this.registryConfigured = true;
    }
  }

  async loginAsRoot() {
    await this.runCommand('rfh auth login --username root --password root1234');
    this.rootJwtToken = this.extractJwtTokenFromConfig();
    this.currentUser = 'root';
  }

  extractJwtTokenFromConfig() {
    // Read JWT token from ~/.rfh/config.toml
    const configPath = path.join(os.homedir(), '.rfh', 'config.toml');
    if (!fs.existsSync(configPath)) {
      return null;
    }
    
    try {
      const config = fs.readFileSync(configPath, 'utf8');
      const parsed = toml.parse(config);
      
      // Try per-registry JWT token first
      if (parsed.current && parsed.registries && parsed.registries[parsed.current]) {
        const registry = parsed.registries[parsed.current];
        if (registry.jwt_token) {
          return registry.jwt_token;
        }
      }
      
      // Fallback to global user JWT token
      if (parsed.user && parsed.user.token) {
        return parsed.user.token;
      }
      
      return null;
    } catch (error) {
      console.error('Failed to parse config.toml for JWT token:', error);
      return null;
    }
  }

  // User creation methods using rfh auth register
  async createTestUser(username, role = 'user') {
    await this.ensureRegistrySetup();
    
    // Check if user already exists in this scenario
    if (this.testUsers.has(username)) {
      return this.testUsers.get(username);
    }
    
    // For this initial implementation, we only support 'user' and 'root' roles
    // All created test users will have 'user' role - use root for write access
    if (role !== 'user' && role !== 'root') {
      throw new Error(`Unsupported role: ${role}. Use 'user' for read-only or 'root' for write access.`);
    }
    
    // Root user already exists, no need to create
    if (role === 'root' && username === 'root') {
      const userData = { username: 'root', role: 'root', email: 'root@rulestack.init' };
      this.testUsers.set(username, userData);
      return userData;
    }
    
    // Use rfh auth register to create test user (always 'user' role)
    try {
      await this.runCommand(
        `rfh auth register --username ${username} --email ${username}@test.com --password testpass123`
      );
      
      const userData = { username, role: 'user', email: `${username}@test.com` };
      this.testUsers.set(username, userData);
      return userData;
      
    } catch (error) {
      // If registration fails (user might already exist), that's usually okay
      // We'll assume the user exists and proceed
      const userData = { username, role: 'user', email: `${username}@test.com` };
      this.testUsers.set(username, userData);
      return userData;
    }
  }

  // Helper method to ensure we have root authentication
  async ensureRootAuthentication() {
    if (this.currentUser !== 'root') {
      await this.loginAsRoot();
    }
  }

  // Role-specific login methods (simplified to user/root only)
  async loginAsUser(username = 'test-user') {
    await this.createTestUser(username, 'user');
    await this.runCommand(`rfh auth login --username ${username} --password testpass123`);
    this.currentUser = username;
  }

  // Helper method for API calls using JWT tokens
  async apiCall(method, endpoint, data = null) {
    if (!fetch) {
      throw new Error('fetch not available - API calls cannot be made');
    }
    
    const baseUrl = 'http://localhost:8081';
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
      }
    };
    
    // Extract current JWT token from config
    const jwtToken = this.extractJwtTokenFromConfig();
    if (jwtToken) {
      options.headers['Authorization'] = `Bearer ${jwtToken}`;
    }
    
    if (data) {
      options.body = JSON.stringify(data);
    }
    
    const response = await fetch(`${baseUrl}${endpoint}`, options);
    return await response.json();
  }

  // Package management methods
  async publishPackage(name, version, content) {
    // Ensure we have root privileges for publishing
    if (this.currentUser !== 'root') {
      await this.loginAsRoot();
    }
    
    // Create package in test-packages directory instead of temp dir
    const packageDir = path.join(__dirname, '../../test-packages', `${name}-${version}`);
    await fs.ensureDir(packageDir);
    
    try {
      // Create rulestack.json manually
      const manifest = {
        name: name,
        version: version,
        description: `Test package for ${name}`,
        targets: ["cursor"],
        tags: ["test"],
        files: ["*.md"],
        license: "MIT"
      };
      
      await fs.writeFile(path.join(packageDir, 'rulestack.json'), JSON.stringify(manifest, null, 2));
      await fs.writeFile(path.join(packageDir, 'rules.mdc'), content);
      
      // Copy config to package directory
      await fs.ensureDir(path.join(packageDir, '.rfh'));
      await fs.copy(path.dirname(this.configPath), path.join(packageDir, '.rfh'));
      
      // Run pack and publish with relative path
      await this.runCommand(`rfh pack rules.mdc --package ${name}`, { cwd: packageDir });
      if (this.lastExitCode !== 0) {
        throw new Error(`Pack failed for ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
      }
      
      await this.runCommand('rfh publish', { cwd: packageDir });
      if (this.lastExitCode !== 0) {
        throw new Error(`Publish failed for ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
      }
    } finally {
      // Clean up package directory
      await fs.remove(packageDir);
    }
  }

  async createPackageDir(name, version, content) {
    const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), `rfh-package-${name}-`));
    
    // Copy the main config to temp directory first (needed for init to work)
    const configDir = path.join(tempDir, '.rfh');
    await fs.ensureDir(configDir);
    
    await fs.copy(path.dirname(this.configPath), configDir);
    if (this.lastExitCode !== 0) {
      throw new Error(`Init failed in temp dir: ${this.lastCommandError || this.lastCommandOutput}`);
    }
    
    // Write rule content
    await fs.writeFile(path.join(tempDir, 'rules.mdc'), content);
    
    return tempDir;
  }

  // Standard test data setup
  async setupTestData() {
    await this.ensureRegistrySetup();
    // TODO: Package publishing setup needs to be fixed
    // For now, just ensure registry is set up - individual tests can handle their own package needs
    console.warn('Package publishing temporarily disabled - tests requiring pre-published packages will be skipped');
  }

  // Verify package exists
  async verifyPackageExists(name, version) {
    const response = await this.apiCall('GET', `/v1/packages/${name}/versions/${version}`);
    return response && !response.error;
  }

  // Enhanced cleanup
  async cleanup() {
    if (this.testDir) {
      process.chdir(this.originalDir);
      await fs.remove(this.testDir);
      this.testDir = null;
    }
    
    // Clear test users map
    if (this.testUsers) {
      this.testUsers.clear();
    }
    
    // Reset authentication state
    this.currentUser = null;
    this.rootJwtToken = null;
    this.registryConfigured = false;
  }

  // Enhanced error handling
  async safeApiCall(method, endpoint, data = null) {
    try {
      return await this.apiCall(method, endpoint, data);
    } catch (error) {
      console.error(`API call failed: ${method} ${endpoint}`, error.message);
      throw new Error(`API call failed: ${error.message}`);
    }
  }

  async safeRunCommand(command, options = {}) {
    try {
      return await this.runCommand(command, options);
    } catch (error) {
      console.error(`Command failed: ${command}`, error.message);
      throw error;
    }
  }

  // Debug logging
  log(message, level = 'info') {
    if (process.env.CUCUMBER_DEBUG) {
      console.log(`[${level.toUpperCase()}] ${message}`);
    }
  }
}

module.exports = CustomWorld;