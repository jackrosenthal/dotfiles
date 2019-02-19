##############################################################################
# Custom keybindings for vi users                                            #
# Will put into vi mode if not already                                       #
##############################################################################

bindkey -v

# vi normal: press q to kill the current line, and receive it on the next
# command
bindkey -M vicmd "q" push-line

## vi normal: press s to insert sudo at the begining of the line
function __zkey_prepend_sudo {
    if [[ $BUFFER != "sudo "* ]]; then
        BUFFER="sudo $BUFFER"
        CURSOR+=5
    fi
}
zle -N prepend-sudo __zkey_prepend_sudo
bindkey -M vicmd "s" prepend-sudo

# Ctrl-R for incremental search backwards (default in emacs, but nice in vi)
stty -ixon
bindkey -M viins '^R' history-incremental-search-backward
bindkey -M viins '^S' history-incremental-search-forward

# Magic dot key
# Autoreprace ... with ../.. and so forth
function __zkey_dot_handler {
    [[ $LBUFFER == *.. ]] && LBUFFER+="/.." || LBUFFER+="."
}
zle -N magic-dot __zkey_dot_handler
bindkey -M viins '.' magic-dot

# Open manual using Shift-H in vi normal mode
autoload -U run-help
bindkey -M vicmd H run-help
