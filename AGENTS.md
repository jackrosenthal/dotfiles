# Dotfiles repo

This is a repo of my personal dotfiles managed by Chezmoi.  It is used to
deploy my personal config across a wide variety of machines.  Right now, I have
it deployed to Arch Linux, Debian, and Ubuntu machines, but this could change.

## Making changes

Make changes directly in this repo, and apply by running `chezmoi apply`.  Do
not partially apply (that is, apply only specific files) unless specifically
requested by the user.

## Profiles

There are three profiles that can be selected during `chezmoi init`.  Take care
to utilize these where needed:

- `laptop`: Has a UI and battery.
- `desktop`: UI only, no battery.
- `lite`: For systems without a UI, e.g., only ever accessed over SSH.

## Desktop Environment

* WM: Sway
* Bar: Waybar
* Display management: Kanshi
* Browser: Google Chrome
* Terminal: Alacritty
* Audio: Pipewire
* Shell: Zsh (note: shell profile is not sourced during DM login)

## Utilities

Utilities get installed in `~/.local/bin`.  The preferred language for
utilities is Go, but small tools can also be shell scripts if they're just a
few lines.

Go utilities get written in `gobin/${toolname}`.  During `chezmoi apply`, these
get built and installed to `~/.local/bin`.  There is no need to manually build
and install these tools, `chezmoi apply` takes care of it.

### Go style

- Code for Go 1.25 or newer
- Prefer `fmt.Sprintf` over string concatenation
- If you need a CLI parser, use Kong
- If you need logging, use `slog`
- If you need fancy UI widgets, use the charm.sh stuff
- Use `gofumpt` for formatting
- Use `golangci-lint` for linting

## Commit style

Start with a verb ("Add", "Fix", "Use"). Sentence case. No conventional commits prefix. No trailing period.

Examples:
- `Add caveman mode badge to claude-statusline`
- `Fix kanshi race on fresh login`
- `Use terminfo tsl/fsl for terminal title setting`

Always include a commit body. The body must:
- Describe what changed and why, not just restate the subject line
- Use complete sentences with explicit subjects (never start a sentence with a bare verb)
- Use consistent tense throughout (present tense is preferred for describing what the code does; past tense is acceptable for describing the previous state before explaining the new behavior)
- Wrap at 72 characters
