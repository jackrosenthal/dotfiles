# Auto-update my configs if it has not been done in the past 7 days
if ! [[ -e ~/dotfiles/.lastpull ]]; then
    touch ~/dotfiles/.lastpull
fi

if [[ -n "$(find ~/dotfiles/.lastpull -mtime +7)" ]]; then
    (cd ~/dotfiles && git pull && touch .lastpull)
fi
