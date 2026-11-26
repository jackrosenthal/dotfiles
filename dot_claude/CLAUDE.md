# User preferences

- Do not agree with or validate a question as if it were a statement.  A
  question is not an assertion; only respond to what was actually said. This
  includes tool-call rejections with a question: answer the question, do not
  silently switch to a different approach.
- Do not invent requirements that were not asked for.
- For complex tasks which can be parallelized, use subagents and act as a
  delegator.
- Do not use build directories to store things that go in the scratchpad.  For
  example, if you write a temporary script you don't want committed, do not put
  it in the `target/` directory in a Rust project.  Build directories should
  only be written in by the official build system.

# Committing

- Commit your work without waiting to be asked, unless told not to.
- Commit directly to the current branch.  Never create a new branch unless explicitly asked.
- When committing work which has not been merged, please amend the appropriate
  commit in the stack when relevant, creating new commits only for distinct work.
- Do not leave untracked files around; add to the `.gitignore` or delete.
- Commit messages should not narrate the diff; assume the reader is able to see
  the code change.
- Describe what **you** changed in an imperative mood.  For example, "Add
  Prometheus metrics on /metrics" instead of "Prometheus metrics serves on
  /metrics".

# English Style

- American English
- Use the oxford comma
- Comma always after e.g., or i.e.,
- Two spaces at to separate sentences
- Avoid usage of the "it" pronoun
- Avoid "own" in possession.  For example, say "The phone's network code",
  instead of "The phone's own network code".
- Avoid "nothing" or "none" in place of a specific claim; say what happens to
  what.  For example:
  - Say "Make only the changes the user asks for", instead of "Change nothing
    the user did not ask for".
  - Say "CI pushes butterfree, but its files are unchanged", instead of "CI
    has nothing to deploy".
  - Say "No more reminders are scheduled", instead of "So none will fire
    again".
- Avoid long-winded headers and descriptions.  For example, instead of saying
  "What a developer needs on their machine" say "Prerequisites".
- When introducing a concept, follow the Define -> Explain -> Example pattern.
- Have empathy for your reader.  Who is your audience?  Why do they want to
  read this?  If what you wrote doesn't match the expectations of the
  audience, delete it.  For example, do not write irrelevant internal
  implementation details in an administration guide.
- Do not mention the software license in a README file if a LICENSE file is
  present.
- Less is more.

# Comment Style

- Only write code comments for non-obvious behavior.
- Do not repeat between comments, if you want to comment something, you get
  one place to say it.
- Only write function or module documentation when the interface is
  non-obvious.
- When writing function or module documentation, describe the interface, not
  the implementation details.
- Less is more.

# Unit Tests

- Unit tests can sometimes be useful, but do not substitute real testing.
- Do not write change detector tests.
- Tests should only test behavior the code exhibits in its branching and logic.
  Do not test something code does not do.  For example, when removing a
  behavior from code, do not write a test that the function no longer does that
  behavior.
- Less is more.

# Code in general

- Unless language style guides say otherwise, prefer string formatting over
  concatenation.  For example, in go, prefer `x := fmt.Sprintf("%s=%s", y, z)`
  to `x := y + "=" + z`.  In Python, use f-strings.
- Do not hand roll something when a library is available to do it, unless
  you've already been asked to hand-roll it.

# Environment

- Use `uv` to build Python environments you require on demand instead of
  relying on what the system Python version is or what packages it has
  available.
- Do not prefer a library version simply because it's already cached on the
  system you're working on.

# Preferred libraries

Go:
- `github.com/alecthomas/kong` for CLI parsing
- `github.com/jackrosenthal/fay` when CLIs will be user facing, to add color
- slog for logging

Rust:
- Clap for CLIs
