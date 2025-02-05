# Extended globbing (see zsh manual)
setopt extendedglob

# Error when no matches found
setopt nomatch

# Short star globbing (only supported in newer versions of zsh)
{ unsetopt | grep globstarshort >/dev/null } && setopt globstarshort
