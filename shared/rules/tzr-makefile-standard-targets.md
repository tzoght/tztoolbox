# Makefile with standard targets

Every project should have a `Makefile` at the repository root that provides a consistent, language-agnostic interface for common development tasks. When creating or modifying a project, ensure these standard targets exist.

## Required targets

| Target | Purpose |
|---|---|
| `help` | Print all available targets with short descriptions (set as `.DEFAULT_GOAL`) |
| `lint` | Run all configured linters / static-analysis tools |
| `fmt` | Auto-format source code (e.g., `black`, `prettier`, `gofmt`) |
| `test` | Run the full test suite |
| `test-unit` | Run only unit tests (fast, no external deps) |
| `test-integration` | Run integration / end-to-end tests |
| `build` | Compile / bundle the project (no-op target if interpreted language) |
| `ci` | Run the full local CI pipeline: `fmt`, `lint`, `test`, `build` — in that order |
| `clean` | Remove generated artifacts, caches, and build output |

## Recommended optional targets

| Target | Purpose |
|---|---|
| `install` | Install the project and/or its dependencies |
| `dev` | Start a local development server or watch mode |
| `docker-build` | Build the container image |
| `docker-run` | Run the container locally |
| `coverage` | Run tests with coverage reporting |
| `check` | Alias for `ci` — some teams prefer this name |

## Conventions

1. **`help` is the default target.** Running bare `make` should print usage, never mutate state.
2. **All targets are `.PHONY`** unless they produce a file of the same name.
3. **`ci` is the pre-PR gate.** Before creating a pull request, the agent (or developer) should run `make ci` and ensure it passes. If `make ci` fails, fix the issues before proceeding.
4. **Targets delegate to the real tooling** (e.g., `lint` calls `ruff check .` or `eslint .`). Keep Makefile recipes thin wrappers so tooling config stays in its own file.
5. **Use variables at the top** of the Makefile for tool paths and flags so they are easy to override:
   ```makefile
   PYTHON   ?= python3
   PIP      ?= pip
   PYTEST   ?= pytest
   RUFF     ?= ruff
   ```
6. **`ci` target should fail fast.** Chain commands with `&&` or use separate recipes so the first failure stops execution.
7. **Add the Makefile early.** When scaffolding a new project, create the Makefile as one of the first files, even with stub targets that print TODOs. Fill them in as tooling is added.

## When to apply

- **New project:** Create the Makefile with all required targets (stubs are fine initially).
- **Existing project without a Makefile:** Propose adding one. Populate targets based on detected tooling (e.g., `package.json` scripts, `pyproject.toml` tools, `Cargo.toml`).
- **Existing project with a Makefile:** Check for missing standard targets and suggest additions. Do not remove existing custom targets.
- **Before creating a PR:** Run `make ci` (or `make check`) and confirm it passes. If the project has no `ci` target, run `make lint && make test` as a fallback.
