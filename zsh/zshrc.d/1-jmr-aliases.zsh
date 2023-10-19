# Functions and aliases that Jack likes

alias pk=pkill
alias g=git
alias xra='xrandr --auto'
alias gnutar='command tar'
alias view='vim -R'
alias ec='emacsclient -n -c -a vim'
alias et='emacsclient -nw -a vim'

if ls --help 2>&1 | grep coreutils >/dev/null; then
    alias ls='ls --color=auto'
    alias la='ls -lah --color=auto'
else
    alias la='ls -lah'
fi

command -v bsdtar >/dev/null && alias tar=bsdtar
command -v python3 >/dev/null && alias python=python3
command -v pip3 >/dev/null && alias pip=pip3
command -v pigz >/dev/null && alias gzip=pigz
command -v unpigz >/dev/null && alias gunzip=unpigz
command -v zathura >/dev/null && alias zathura='zathura --fork'
command -v bat >/dev/null && alias cat='bat'
command -v batcat >/dev/null && alias bat='batcat' && alias cat='batcat'

# make a directory (-p) and cd to it
function take () {
    mkdir -p "$1" && cd "$1"
}

# Run whatever command fully upgrades a system
function sysup {
    if [[ -f /etc/apt/sources.list ]]; then
        sudo apt update && sudo apt dist-upgrade && sudo apt autoremove
    elif [[ -f /etc/pacman.conf ]]; then
        sudo pacman -Syu
    fi
}

# suffix aliases
alias -s pdf="${PDFVIEW:=zathura}"
alias -s tex="${EDITOR}"
alias -s png=feh
alias -s jpg=feh
alias -s jpeg=feh
alias -s bmp=feh
alias -s ppm=feh
alias -s gif=feh
alias -s xbm=feh
