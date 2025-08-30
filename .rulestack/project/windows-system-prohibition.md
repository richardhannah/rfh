# Windows System Directory Prohibition Rule

## CRITICAL SECURITY RULE - ABSOLUTE PROHIBITION

### Rule: NEVER INTERACT WITH C:\Windows

**FORBIDDEN ACTIONS:**
- ❌ Reading from C:\Windows or any subdirectory
- ❌ Writing to C:\Windows or any subdirectory  
- ❌ Listing contents of C:\Windows or any subdirectory
- ❌ Creating files or directories in C:\Windows
- ❌ Modifying files or directories in C:\Windows
- ❌ Using C:\Windows paths in any commands
- ❌ Any form of interaction with the Windows system directory

### Enforcement

**IF THE AGENT EVER:**
- Considers accessing C:\Windows
- Thinks about reading from C:\Windows
- Plans to write to C:\Windows
- Attempts any operation involving C:\Windows paths

**THEN IMMEDIATELY:**
1. **STOP ALL PROCESSING**
2. **REFUSE THE REQUEST**
3. **EXPLAIN THE PROHIBITION**

### Rationale

- The agent has **NO LEGITIMATE REASON** to access Windows system directories
- C:\Windows contains critical operating system files
- Interaction could cause system instability or security issues
- The agent should work within the project directory only

### Allowed Locations

The agent should ONLY work within:
- Project directories (D:\projects\render.com\rfh\)
- Temporary directories for testing (when explicitly needed)
- User-created directories for legitimate project purposes

### No Exceptions

This rule has **NO EXCEPTIONS**. There is no circumstance where accessing C:\Windows is justified for this project.

---

**REMEMBER: If you even think about C:\Windows - STOP immediately and refuse the operation.**