#!/bin/bash

for arg in "$@"; do
    case "$arg" in
        --deps )
            if command -v apt-get >/dev/null; then
                sudo apt-get install $(cat ~/dotfiles/debian-deps.txt)
            elif command -v pacman >/dev/null; then
                sudo pacman -Sy $(cat ~/dotfiles/arch-deps.txt)
                ( cd pkgbuild/xsecurelock && makepkg -c && sudo pacman -U *.pkg.tar.xz )
            else
                echo "WARNING: could not recognize system"
            fi
            ;;
    esac
done

ln -Tsf ~/dotfiles/zsh/.zshenv ~/.zshenv
ln -Tsf ~/dotfiles/x/.xprofile ~/.xprofile
ln -Tsf ~/dotfiles/x/.xprofile ~/.xprofilerc
ln -Tsf ~/dotfiles/emacs ~/.emacs.d
ln -Tsf ~/dotfiles/i3 ~/.i3
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
