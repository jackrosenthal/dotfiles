# Nice completion
autoload -Uz compinit && compinit

# Menu highlighting
command -v dircolors >/dev/null && source <(dircolors | sed -E 's/\bsetenv (\w+) /export \1=/')
zstyle ':completion:*' list-colors ${(s.:.)LS_COLORS}

# Automatic rehash
zstyle ':completion:*' rehash true

zstyle ':compinstall' filename ~/.zshrc
