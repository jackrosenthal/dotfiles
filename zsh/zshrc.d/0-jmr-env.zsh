# Jack's preferred environment variables
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
