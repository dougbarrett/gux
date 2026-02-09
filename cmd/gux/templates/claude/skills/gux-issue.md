---
name: gux:issue
description: File a new gux framework issue or pull in an existing issue to work on
argument-hint: "<new|issue-number>"
---

# Gux Issue: $ARGUMENTS

You are managing a GitHub issue for the dougbarrett/gux framework repository. The argument determines the mode:

- If `$ARGUMENTS` is `new` → File a new issue (Mode 1)
- If `$ARGUMENTS` is a number → Work on that existing issue (Mode 2)

---

## Mode 1: File a New Issue (`new`)

### Step 1: Gather information

Ask the user to describe the problem or feature request. Prompt for:
- What they expected vs what happened
- Steps to reproduce (if a bug)
- Which part of the framework is affected (core, codegen, CLI, UI, etc.)

### Step 2: Investigate context

Use Grep, Glob, and Read to find relevant code in the gux codebase that relates to the described issue. Identify:
- Which files and functions are involved
- Related existing tests
- Whether this is a known pattern or a new area

### Step 3: Get gux version

Run `gux version` to capture the current version for the issue report.

### Step 4: Draft the issue

Auto-generate a structured GitHub issue with:
- **Title**: Concise, descriptive (e.g., "codegen: incorrect pointer handling in DTO field mapping")
- **Body** with sections:
  - **Description**: Clear summary of the problem or feature
  - **Steps to Reproduce** (if bug): Numbered steps
  - **Expected Behavior**: What should happen
  - **Actual Behavior** (if bug): What currently happens
  - **Environment**: gux version from step 3
  - **Relevant Code**: File paths and line numbers identified in step 2

Show the draft to the user for approval before creating.

### Step 5: Create the issue

```bash
gh issue create --repo dougbarrett/gux --title "<title>" --body "$(cat <<'EOF'
<body content>
EOF
)"
```

Return the issue URL to the user.

---

## Mode 2: Work on Existing Issue (`<number>`)

### Step 1: Fetch the issue

Run `gh issue view $ARGUMENTS --repo dougbarrett/gux --json title,body,labels,comments,url` to get the full issue details. Display the title, body, and any comments to understand the requirements.

### Step 2: Update gux to latest version

Run `gux update` to ensure we're working with the latest gux binary.

### Step 3: Restart the dev server

1. Check for any running `gux dev` process: `pgrep -f 'gux dev'`
2. If running, kill it: `pkill -f 'gux dev'`
3. Start the dev server in the background: run `gux dev` using background execution
4. Wait for the server to be ready (check for the listening message)

### Step 4: Investigate the codebase

- Use Grep, Glob, and Read to locate the relevant source files
- Understand the existing code, patterns, and architecture
- Check related test files to understand testing patterns

### Step 5: Implement the fix or feature

- Make the minimal changes needed
- Follow existing code patterns and style
- Do not refactor unrelated code or add unnecessary features

### Step 6: Write tests

- Add unit tests that verify the fix/feature works
- Add regression tests that would have caught the original bug
- Follow existing test patterns (table-driven tests with `t.Run()` subtests, `t.TempDir()` for filesystem tests)
- Reset `modelFieldTypesCache` in codegen tests that use `getModelFieldTypes()`

### Step 7: Run unit tests

Run `go test ./cmd/gux/... ./core/... ./ui/... -count=1 -timeout 120s` to verify all tests pass. Fix any failures before proceeding.

### Step 8: Run Playwright MCP regression tests

Use the Playwright MCP browser tools to verify the app works end-to-end:

1. **Check for test scripts**: Glob for `*.test.ts`, `*.spec.ts`, `e2e/**`, `tests/**` and run any existing Playwright test scripts found
2. **Navigate key pages**: Use `browser_navigate` to visit critical pages (home, admin dashboard, relevant model pages)
3. **Take screenshots**: Use `browser_take_screenshot` on each critical page to visually verify rendering
4. **Check console errors**: Use `browser_console_messages` with level `error` to ensure no JavaScript errors
5. **Verify the fix**: Navigate to the specific page or feature affected by the issue and confirm the fix works

### Step 9: Commit

- Stage only the changed files by name (not `git add .`)
- Write a commit message in this format:
  ```
  fix(<scope>): <short description> (#$ARGUMENTS)

  <longer explanation of what was wrong and how this fixes it>

  Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
  ```
- Use the appropriate scope: `codegen`, `crud`, `ui`, `build`, `cli`, `core`, etc.
- For new features use `feat(<scope>)` instead of `fix(<scope>)`

### Step 10: Push to GitHub

Run `git push origin main`.

### Step 11: Tag a new version

- Check the latest tag with `git tag --sort=-v:refname | head -1`
- Increment the patch version (e.g., v1.28.66 -> v1.28.67)
- Create and push the tag: `git tag <new-version> && git push origin <new-version>`

### Step 12: Create a GitHub release

Use `gh release create` with:
- Title: `<version> - <short description of fix>`
- Body: Markdown with `## Bug Fix` or `## New Feature` header, root cause explanation, what was fixed, and list of tests added
- Use a HEREDOC for the body to preserve formatting

### Step 13: Comment on the issue

Use `gh issue comment $ARGUMENTS --repo dougbarrett/gux` with:
- Link to the release
- Brief explanation of the fix
- List of tests added
