# Jack's preferred environment variables

prepend_path() {
    export PATH="$1:${PATH}"
}

append_path() {
    export PATH="${PATH}:$1"
}

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

: ${EDITOR:=$(pref-order vim vi)}
: ${PAGER:=$(pref-order less more)}
: ${PDFVIEW:=$(pref-order zathura evince mupdf xpdf)}
export EDITOR PAGER PDFVIEW
export CLICOLOR=t

umask 022

if [[ -z "${SSH_AUTH_SOCK}" && -S "${XDG_RUNTIME_DIR}/ssh-agent.sock" ]]; then
    export SSH_AUTH_SOCK="${XDG_RUNTIME_DIR}/ssh-agent.sock"
fi
