# AGENTS.md

## Conventions

- Do not run git add, commit or push. You can suggest good times to run those commands, but do not run them yourself. If you want to run git add, commit or push, ask for permission first.
- When writing summaries, aim to be concise. Do not include long blocks of code examples.
- Do not write summaries to disk without asking permission first.
- When adding a feature or any new piece of testable code, remember to add a unit test. If possible, start by writing the test first, then implement the feature to make the test pass.
- Unit tests should be isolated from each other.
- Unit tests should avoid mocks and stubs. If these seem needed, consider refactoring the code to be more testable instead - but ask permission before making any significant refactorings.
- Unit tests should test one unit of code at a time, and should not have side effects that impact other tests.
- Avoid making functions public just to test them.
