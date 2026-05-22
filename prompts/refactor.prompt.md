# Refactoring Analysis Prompt

Analyze the refactoring of `{TARGET_MODULE}`.

## Current problem

{DESCRIPTION_OF_ISSUE}

## What to analyze

1. What is the current structure and what makes it hard to change?
2. What are 2-3 possible approaches to restructure?
3. What is the blast radius (files affected, tests to update)?
4. Which approach best aligns with the project's architecture (`docs/architecture/`)?

## Constraints

- Do not propose adding new dependencies
- Do not change public interfaces unless necessary
- Must maintain backward compatibility with existing game protocol (WebSocket message format)

## Output format

```
### Current Structure
{brief description of current code organization}

### Approaches
1. **{Approach A}**: {description, files affected, risk level}
2. **{Approach B}**: {description, files affected, risk level}
3. **{Approach C}**: {description, files affected, risk level}

### Recommendation
{which approach and why}
```
