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
    
    // Configuration paths - use isolated test config, not user's real config
    this.testConfigDir = null; // Will be set in createTempDirectory
    this.configPath = null; // Will be set based on testConfigDir
    
    // Enhanced World properties for authentication and registry management
    this.rootJwtToken = null;
    this.currentUser = null;
    this.registryConfigured = false;
    this.testUsers = new Map(); // Track created test users
  }

  async createTempDirectory() {
    this.testDir = await fs.mkdtemp(path.join(os.tmpdir(), 'rfh-test-'));
    // Store the current working directory and change to test directory
    this.tempProjectDir = this.testDir;  // Alias for backward compatibility
    
    // Set up shared cucumber config directory (production-like)
    this.testConfigDir = path.join(os.homedir(), '.rfh-cucumber');
    this.configPath = path.join(this.testConfigDir, 'config.toml');
    await fs.ensureDir(this.testConfigDir);
    
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
      // Ensure we have consistent config directory path
      const configDir = this.testConfigDir;
      if (!configDir) {
        throw new Error('Test config directory not initialized. Call createTempDirectory first.');
      }
      await fs.ensureDir(configDir);
      
      // Replace 'rfh' with the actual binary path (handle both 'rfh ' and 'rfh' at end of string)
      const actualCommand = command.replace(/^rfh(\s|$)/, `"${this.rfhBinary}"$1`);
      
      // Set up environment with RFH_CONFIG for test isolation
      const env = {
        ...process.env,
        RFH_CONFIG: configDir,
        ...options.env
      };
      
      // Environment setup complete - RFH_CONFIG is now set for test isolation
      
      this.lastCommandOutput = execSync(actualCommand, {
        cwd: options.cwd || this.testDir || process.cwd(),
        encoding: 'utf8',
        timeout: 30000,
        env: env
      });
      this.lastExitCode = 0;
      this.lastCommandExitCode = 0;
    } catch (error) {
      this.lastCommandError = error.stderr || error.message || '';
      this.lastCommandOutput = error.stdout || '';
      this.lastExitCode = error.status || 1;
      this.lastCommandExitCode = error.status || 1;
      
      // Debug logging for failed execution
      this.log(`Command failed with exit code: ${this.lastExitCode}`, 'error');
      this.log(`Error: ${this.lastCommandError}`, 'error');
      this.log(`Stdout: ${this.lastCommandOutput}`, 'error');
    }
  }

  async runCommandInDirectory(command, directory, options = {}) {
    try {
      // Ensure we have consistent config directory path
      const configDir = this.testConfigDir;
      if (!configDir) {
        throw new Error('Test config directory not initialized. Call createTempDirectory first.');
      }
      await fs.ensureDir(configDir);
      
      // Replace 'rfh' with the actual binary path (handle both 'rfh ' and 'rfh' at end of string)
      const actualCommand = command.replace(/^rfh(\s|$)/, `"${this.rfhBinary}"$1`);
      
      // Set up environment with RFH_CONFIG for test isolation
      const env = {
        ...process.env,
        RFH_CONFIG: configDir,
        ...options.env
      };
      
      // Environment setup complete - RFH_CONFIG is now set for test isolation
      
      this.lastCommandOutput = execSync(actualCommand, {
        cwd: directory,
        encoding: 'utf8',
        timeout: 30000,
        env: env
      });
      this.lastExitCode = 0;
      this.lastCommandExitCode = 0;
    } catch (error) {
      this.lastCommandError = error.stderr || error.message || '';
      this.lastCommandOutput = error.stdout || '';
      this.lastExitCode = error.status || 1;
      this.lastCommandExitCode = error.status || 1;
      
      // Debug logging for failed execution
      this.log(`Command failed with exit code: ${this.lastExitCode}`, 'error');
      this.log(`Error: ${this.lastCommandError}`, 'error');
      this.log(`Stdout: ${this.lastCommandOutput}`, 'error');
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
    
    // Create a temporary isolated directory for this package
    const packageDir = await fs.mkdtemp(path.join(os.tmpdir(), `rfh-publish-${name}-`));
    
    try {
      // Initialize RFH in the temp directory first
      await this.runCommand('rfh init --package', { cwd: packageDir });
      if (this.lastExitCode !== 0) {
        throw new Error(`Init failed for ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
      }
      
      // Copy authentication from test config directory (not the user's real config)
      const packageConfigDir = path.join(packageDir, '.rfh-package-config');
      
      // Copy the test config if it exists
      if (this.testConfigDir && await fs.pathExists(this.testConfigDir)) {
        await fs.copy(this.testConfigDir, packageConfigDir);
        console.log(`Copied test config from ${this.testConfigDir} to ${packageConfigDir}`);
      } else {
        // Create empty config dir if test config doesn't exist yet
        await fs.ensureDir(packageConfigDir);
      }
      
      // Create the rule content file
      await fs.writeFile(path.join(packageDir, 'rules.mdc'), content);
      
      // Update the generated rulestack.json with correct package info
      const manifestPath = path.join(packageDir, 'rulestack.json');
      let manifest;
      if (await fs.pathExists(manifestPath)) {
        const manifestContent = await fs.readFile(manifestPath, 'utf8');
        manifest = JSON.parse(manifestContent);
        // Update the first entry if it's an array, or the object itself
        if (Array.isArray(manifest)) {
          manifest[0].name = name;
          manifest[0].version = version;
          manifest[0].description = `Test package for ${name}`;
          manifest[0].files = ["rules.mdc"];
        } else {
          manifest.name = name;
          manifest.version = version;
          manifest.description = `Test package for ${name}`;
          manifest.files = ["rules.mdc"];
        }
      } else {
        // Create manifest as array format for package mode
        manifest = [{
          name: name,
          version: version,
          description: `Test package for ${name}`,
          targets: ["cursor"],
          tags: ["test"],
          files: ["rules.mdc"],
          license: "MIT"
        }];
      }
      
      await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2));
      
      // Save current directory and change to package directory manually
      const originalDir = process.cwd();
      try {
        process.chdir(packageDir);
        
        // Use the package-specific config directory for all commands
        const packageEnv = { RFH_CONFIG: packageConfigDir };
        
        // Pack the package from the correct directory with isolated config
        await this.runCommand(`rfh pack rules.mdc --package ${name}`, { env: packageEnv });
        if (this.lastExitCode !== 0) {
          throw new Error(`Pack failed for ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
        }
        
        // Ensure authentication in the package directory context
        await this.runCommand('rfh auth login --username root --password root1234', { env: packageEnv });
        if (this.lastExitCode !== 0) {
          console.warn(`Auth login failed for package ${name}: ${this.lastCommandOutput}`);
        }
        
        // Publish the package
        await this.runCommand('rfh publish', { env: packageEnv });
        if (this.lastExitCode !== 0) {
          throw new Error(`Publish failed for ${name}: ${this.lastCommandError || this.lastCommandOutput}`);
        }
        
      } finally {
        // Always restore original directory
        process.chdir(originalDir);
      }
      
      console.log(`Successfully published ${name}@${version}`);
      
    } finally {
      // Clean up package directory
      try {
        await fs.remove(packageDir);
      } catch (cleanupError) {
        console.warn(`Failed to cleanup package dir: ${cleanupError.message}`);
      }
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
    
    // Publish test packages that are expected by the tests
    try {
      await this.publishPackage('security-rules', '1.0.1', '# Security Rules v1.0.1\n\nTest security rules for testing purposes.');
      await this.publishPackage('example-rules', '0.1.0', '# Example Rules v0.1.0\n\nTest example rules for testing purposes.');
      console.log('Test packages published successfully');
    } catch (error) {
      console.warn(`Warning: Failed to publish test packages: ${error.message}`);
      // Don't throw - tests might still work if packages are already on server
    }
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