# AI Agent Workflow Instructions

Whenever you are assigned a task or asked to implement a feature, strictly adhere to the following workflow:

## Task Execution Protocol
1. **Task Selection:** Pick a new task from Phase 1 in Linear.
2. **Implementation:** Write clean, production-ready code fulfilling the requirements.
3. **Validation:** Run local builds and verify that the changes work as expected.
4. **Testing:** Add or update unit/integration tests whenever relevant.
5. **Commit:** Create a Git commit following the strict formatting rules below.
6. **Close the task:** Once the commit is made and everything passes, mark the Linear task as Done.

## Commit Message Standards
Your commit messages MUST follow this exact structure:

[<category>] <small description of the feature>

<Extended message line 1 describing WHY the change was made (max 180 chars)>
<Extended message line 2 describing WHAT was changed (max 180 chars)>

### Category Options
- `feat` (new feature)
- `fix` (bug fix)
- `test` (adding or updating tests)
- `refactor` (code restructuring without behavioral changes)
- `docs` (documentation updates)
- `chore` (maintenance, dependencies, build scripts)

### Commit Example
[feat] add user authentication handler

Implement JWT validation middleware to protect API endpoints from unauthenticated access.
Update router configuration to inject auth context into request handlers and write integration tests.

## Code Comment Standards
A comment block MUST NOT exceed 2 lines, ever. Each of those lines MUST NOT exceed 120 characters.
That is a hard cap of 2 newline characters (`\n`) per comment block - not 2 sentences, not 2 paragraphs: 2 physical lines, full stop.
Comments MUST NOT reference issue IDs (e.g. Linear/Jira tickets like SKO-93).

## Testing Standards
Tests MUST use `github.com/stretchr/testify/require` for assertions (e.g. `require.NoError`, `require.Equal`, `require.Contains`) instead of manual `if`/`t.Fatalf`/`t.Errorf` checks.
