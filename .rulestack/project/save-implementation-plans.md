# Save Implementation Plans Rule

## Rule
**BEFORE** starting implementation work, save the approved plan to the `planning/` folder.

## Process
1. Create implementation plan and get user approval
2. Save the plan as a markdown file in `planning/` directory
3. Use descriptive filename: `[feature-name]-implementation.md`
4. Then proceed with implementation

## Planning File Structure
```
planning/
├── feature-name-implementation.md
├── another-feature-implementation.md
└── bug-fix-implementation.md
```

## Benefits
- Creates permanent record of implementation decisions
- Allows referencing original plan during development
- Provides context for future developers
- Documents the "why" behind implementation choices
- Enables plan review and improvement over time

## File Content Format
Include in the planning file:
- Brief description of the feature/change
- Implementation approach decided upon
- Key files to be created/modified
- Testing strategy
- Any important considerations or constraints

## When to Create
- New features
- Significant refactoring
- Complex bug fixes
- Architecture changes
- Any multi-step implementation