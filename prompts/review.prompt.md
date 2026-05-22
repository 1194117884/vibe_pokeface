# Code Review Prompt

Review the changes on branch `{BRANCH}` against `main`.

## What to check

1. **Correctness**: Does the code do what it claims? Are edge cases handled?
2. **Safety**: No SQL injection, XSS, token leaks, or exposed secrets.
3. **Architecture**: Changes follow existing patterns in `docs/architecture/`. No new patterns without justification.
4. **Types**: No `any` in TypeScript. Proper Go error wrapping.
5. **Tests**: New game logic has tests. No broken existing tests.
6. **Scope**: No unrelated changes. No mass refactoring.

## What NOT to do

- Do not suggest adding features beyond the PR scope
- Do not demand test coverage for trivial UI changes
- Do not suggest dependency upgrades unless security-critical

## Output format

```
### Review Summary
{2-3 sentence overall assessment}

### Issues Found
- [ ] **Critical**: {issue} (must fix)
- [ ] **Major**: {issue} (should fix)
- [ ] **Minor**: {issue} (nice to fix)

### Verified
- [ ] Type safety
- [ ] Error handling
- [ ] Test coverage for changes
- [ ] No security issues
```
