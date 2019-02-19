setopt notify
setopt nobeep
setopt autopushd
setopt autocd

setopt extendedglob
for f in ~/dotfiles/zsh/zshrc.d/*; do
    source "$f"
done
