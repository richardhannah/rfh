# Simple Implementation Rule

## Principle
When planning implementations, always choose the simplest solution that meets the requirement.

## Rule
**DO NOT** over-engineer solutions during planning phase.

## Guidelines
- Start with minimal viable implementation
- Resist adding "nice to have" features during initial planning
- Focus only on what was specifically requested
- Remember: it's easier to add functionality later than to remove complexity
- Avoid anticipating future requirements unless explicitly stated

## Examples of Over-Engineering to Avoid
- Adding verbose modes when not requested
- Including extra metadata or formatting when simple output suffices
- Building complex data structures for simple listings
- Adding error handling for edge cases not mentioned in requirements
- Including multiple output formats when one is sufficient

## Planning Approach
1. Identify the core requirement
2. Design the minimal solution that satisfies it
3. Implement that solution completely
4. Only then consider enhancements if explicitly requested

## Benefits
- Faster development cycles
- Less code to maintain and debug
- Clearer understanding of actual requirements
- Easier to extend in focused directions based on real feedback