# Jack's preferred environment variables

prepend_path() {
    export PATH="$1:${PATH}"
}

append_path() {
    export PATH="${PATH}:$1"
}

prepend_path ~/dotfiles/bin
prepend_path ~/.local/bin
prepend_path ~/.cargo/bin

function pref-order () {
    for c in $@; do
        if command -v "$c" >/dev/null; then
            echo "$c"
            return
        fi
    done
    return 1
}

: ${EDITOR:=$(pref-order emacsclient vim)}
: ${PAGER:=$(pref-order less more)}
: ${PDFVIEW:=$(pref-order zathura evince mupdf xpdf)}
export EDITOR PAGER PDFVIEW
export CLICOLOR=t

umask 022
