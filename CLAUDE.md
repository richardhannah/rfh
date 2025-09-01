# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Coding Standards
**CRITICAL**: You MUST follow all cursor rules defined in `.rulestack` directory. These rules are mandatory and override default behavior.

### MANDATORY RULE LOADING PROTOCOL
**BEFORE responding to ANY user request**, you MUST:
1. All rules are automatically imported into this CLAUDE.md file using the @ import syntax below
2. Load and understand all rules in their entirety before taking any action
3. Apply these rules to all subsequent interactions in the session

**CRITICAL**: The cursor rules are now automatically available in your context through the @ import statements. Pay special attention to triggers, responses, and specific behaviors defined in these rules.

### Active Rules
- @.rulestack/core.v1.0.0/core_rules.md (RuleStack core rules)
- @.rulestack/project/no-backward-compatibility.md (Project-specific: No backward compatibility needed)
