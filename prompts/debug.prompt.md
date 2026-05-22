# Debug Investigation Prompt

Investigate the bug: `{BUG_DESCRIPTION}`

## Context

- **Observed behavior**: {WHAT_HAPPENS}
- **Expected behavior**: {WHAT_SHOULD_HAPPEN}
- **Reproduction steps**: {STEPS}
- **Relevant files**: {FILE_PATHS}

## Investigation process

1. Read the relevant code paths end-to-end
2. Identify where the observed behavior diverges from expected
3. Check: recent changes (`git log --oneline -10`), data flow, edge cases
4. Propose a root cause with evidence (not guesses)

## What NOT to do

- Do not propose a fix without identifying root cause
- Do not make changes outside the affected code path
- Do not refactor "while you're here"

## Output format

```
### Root Cause
{concise explanation of what's wrong and why}

### Evidence
- {file:line} — {what the code does vs. what it should do}
- {additional evidence}

### Fix Plan
1. {specific change in file X}
2. {specific change in file Y}

### Verification
- How to confirm the fix works
```
