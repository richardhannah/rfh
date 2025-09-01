# RFH Test Runners

This directory contains multiple scripts to run the Cucumber BDD tests for RFH.

## Available Scripts

### 🪟 Windows
- **`run-tests.ps1`** - PowerShell script (recommended)

### 🐧 Linux/Mac  
- **`run-tests.sh`** - Bash script

## Usage

### PowerShell (Windows)
```powershell
# Run working tests only (recommended)
./run-tests.ps1 actual

# Run all init tests  
./run-tests.ps1 init

# Run all tests (may have failures)
./run-tests.ps1 all
```

### Bash (Git Bash/WSL)
```bash
# Make executable (first time only)
chmod +x run-tests.sh

# Run working tests only
./run-tests.sh actual

# Run all init tests
./run-tests.sh init
```

## Test Categories

- **`actual`** - Only tests that currently pass (validates working features)
- **`init`** - All rfh init related tests (may include failing tests for unimplemented features)
- **`all`** - All available tests (includes many failing tests)

## Requirements

- **Node.js** (v16+ recommended)
- **npm** (comes with Node.js)
- **Go** (to build RFH binary)

## What the Scripts Do

1. ✅ Verify Node.js and Go are available
2. 🏗️ Build the latest RFH binary (`dist/rfh.exe`)  
3. 📦 Install npm dependencies (if needed)
4. 🧪 Run Cucumber tests with specified target
5. 📊 Display results and exit with appropriate code

## Expected Results

**`actual` target should show:**
- ✅ 3-4 scenarios passing
- Basic rfh init functionality working
- Scope removal working correctly

**`init` and `all` targets will show:**
- ❌ Several failing tests (expected - features not implemented)  
- ⚠️ Many undefined step definitions
- ✅ Some passing core functionality

## Troubleshooting

**"Node.js not found"**: Install from https://nodejs.org/

**"RFH binary build failed"**: Ensure Go is installed and you're in the RFH root directory

**"Permission denied"**: On Unix systems, run `chmod +x run-tests.sh` first

**Tests running slowly**: This is normal - tests create temporary directories and run real RFH commands