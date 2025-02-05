#!/bin/bash

if [[ -d ~/dotfiles ]]; then
  rm -f \
    ~/.config/kitty \
    ~/.config/rofi \
    ~/.emacs.d \
    ~/.i3 \
    ~/.ssh/config \
    ~/.vimrc \
    ~/.xprofile \
    ~/.xsessionrc \
    ~/.zshenv

  rm -rf ~/dotfiles
fi
