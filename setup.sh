#!/bin/bash

for arg in "$@"; do
    case "$arg" in
        --deps )
            if command -v apt-get >/dev/null; then
                sudo apt-get install $(cat ~/dotfiles/debian-deps.txt)
            elif command -v pacman >/dev/null; then
                sudo pacman -Sy --needed $(cat ~/dotfiles/arch-deps.txt)
            else
                echo "WARNING: could not recognize system"
            fi
            ;;
    esac
done

mkdir -p ~/.config
ln -Tsf ~/dotfiles/zsh/.zshenv ~/.zshenv
ln -Tsf ~/dotfiles/x/.xprofile ~/.xprofile
ln -Tsf ~/dotfiles/x/.xprofile ~/.xsessionrc
ln -Tsf ~/dotfiles/emacs ~/.emacs.d
ln -Tsf ~/dotfiles/i3 ~/.i3
ln -Tsf ~/dotfiles/kitty ~/.config/kitty
ln -Tsf ~/dotfiles/rofi ~/.config/rofi
ln -Tsf ~/dotfiles/.vimrc ~/.vimrc

mkdir -p ~/.local/vim{swap,undo}

if [[ -e ~/.emacs ]]; then
    mv ~/.emacs ~/.emacs.old
    echo "WARNING: renamed ~/.emacs -> ~/.emacs.old"
fi

if command -v systemctl >/dev/null && systemctl status --user emacs >/dev/null; then
    systemctl enable --user emacs
else
    echo "WARNING: emacs service not enabled automatically"
fi

poetry -C ~/dotfiles/3rdparty/rofimoji install
