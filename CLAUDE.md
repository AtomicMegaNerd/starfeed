# LLM Policies

## General Policies

- This project uses a Nix flake with nix-direnv. You need to enable the flake environment for
  the Go toolchain to work.
- Always run `task build`, `task test`, and `task lint` after making any changes to the code.
- Always ensure each line is less than 100 characters long regardless of the file type.
- Only the human is allowed to make changes to the README.md file or the CLAUDE.md file.

## Git Policies

- When the human asks you to make a commit, always create a new branch named
  `feature/short-description-of-change` and make the commit there.
- When the human asks you to commit, they are giving you explicit permission to make
  the commit.

## Go Policies

- Keep the code as simple as possible.
- Use guards and short-circuiting to avoid deep nesting.
- Use custom error types if it makes the code simpler to read.
- Always handle errors explicitly.
- Always use context in functions that make network calls or do I/O.
- Always use dependency injection for anything that does I/O or network calls.
- Warn me if a function or a file will get too big. We can split it up.
