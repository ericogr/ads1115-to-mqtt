Configuration file for CODEX agents

Scope
- This file applies to the entire repository (root).

Purpose
- Provide basic guidelines for automated agents (e.g., CODEX) working in this repository.

Project environment
- Go project using modules (`go.mod`).
- Main structure: `main.go`, `pkg/`, `.github/workflows/`.

Allowed actions
- Read any file in the repository.
- Run analysis and test commands that do not modify the repository, for example: `go mod download`, `go test ./...`, `go build`.
- Apply patches/edits via `apply_patch` (or an equivalent mechanism) to propose changes.

Important restrictions (READ BEFORE ACTING)
- DO NOT run `git` operations that modify history, the index, or the remote. Examples of prohibited commands:
  - `git add`, `git commit`, `git push`, `git pull`, `git fetch --all`
  - `git reset --hard`, `git rebase`, `git merge`, `git checkout -B`
  - `git remote add` or any commands that change branches/remotes
- If a change is necessary, create a patch (`apply_patch`) and describe the change in the comment; request that the maintainer/user review and perform the commit/push manually.

Approvals and sensitive actions
- For any action that requires elevated permissions, network access, or that may be destructive (e.g., `rm -rf`, changes outside the workspace folder), ask for explicit user approval before executing.

Best practices
- Keep changes minimal and focused on the requested issue.
- Follow the existing code style in the repository and do not rename files without a clear reason.
- Do not add automatic commits; propose patches and wait for confirmation.

Validation
- Whenever possible, run `go build` and `go test` locally and report the results to the user.

Contact / Recommended flow
- Propose changes with `apply_patch` and wait for human review before commits and pushes.
