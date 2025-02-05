##############################################################################
# Jack's prompt line                                                         #
##############################################################################

autoload -U colors && colors
autoload -U is-at-least

if [[ -n $SSH_CONNECTION ]]; then
    # We are on an SSH connection
    SSH_PROMPT='%{%B$fg[yellow]%}[SSH: %m]%{$reset_color%}%b '
fi

function __zhook_return_code {
    local exit_status="$?"
    if [[ "${exit_status}" -eq 0 ]]; then
        RC_DISPLAYED=false
    elif ! "${RC_DISPLAYED}"; then
        echo -n "${fg[red]}"
        echo ">>> Command exited with status ${exit_status}"
        echo -n "${reset_color}"
        RC_DISPLAYED=true
    fi
}

precmd_functions=(__zhook_return_code)

function __zprompt_mode {
    case ${KEYMAP} in
        (vicmd) echo "%{$fg_bold[blue]%}N%{$reset_color%}" ;;
        (visual) echo "%{$fg_bold[blue]%}V%{$reset_color%}" ;;
        (*) echo "%#" ;;
    esac
}

function __zprompt_env {
    if [[ -n "${VIRTUAL_ENV}" ]]; then
        echo "%{$fg_bold[blue]%}[ENV: ${VIRTUAL_ENV:t}]%{$reset_color%} "
    fi
}

setopt prompt_subst # allow running functions in the prompt

function set-prompt {
    PROMPT="${SSH_PROMPT}"
    PROMPT+='$(__zprompt_env)'
    PROMPT+='%{$fg[blue]%}%n%{$fg[green]%} %~%{$reset_color%} '
    PROMPT+='$(git_super_status)$(__zprompt_mode) '
}

if is-at-least 5.1; then
    # bind edit mode display
    function zle-line-init zle-keymap-select () {
        set-prompt
        zle reset-prompt
    }
    zle -N zle-line-init
    zle -N zle-keymap-select
else
    set-prompt
fi

if ! command -v git_super_status >/dev/null; then
    git_super_status () { echo; }
fi
