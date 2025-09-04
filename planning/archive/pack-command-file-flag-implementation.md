# Pack Command File Flag Implementation Plan

## Feature Description
Change pack command from using positional argument for file input to using the `--file` flag instead.

## Current vs New Behavior
**Current:** `rfh pack my-rule.mdc --package=test-package`  
**New:** `rfh pack --file=my-rule.mdc --package=test-package`

## Implementation Approach
1. Update command definition to remove positional argument requirement
2. Make `--file` flag required and use `fileOverride` variable
3. Update function signatures to use flag variable instead of positional argument
4. Keep all other pack logic unchanged

## Key Files to Modify
1. **`internal/cli/pack.go`** - Update command definition and argument handling
2. **`internal/cli/pack_workflows.go`** - Update function calls to use `fileOverride`

## Expected New Usage
```
Usage:
  rfh pack [flags]

Flags:
  -f, --file string      .mdc file to pack (required)
  -p, --package string   package name (enables non-interactive mode)  
      --version string   package version (default: 1.0.0)
```

## Key Constraints
- Keep all existing pack functionality intact
- Only change how file input is specified (positional → flag)
- Maintain interactive and non-interactive modes

## Testing Strategy
- Test both interactive and non-interactive modes with new flag
- Verify error handling when `--file` is missing
- Update any existing tests that use positional argument

## Implementation Status
Ready to implement