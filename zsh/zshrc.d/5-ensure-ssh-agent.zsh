if [[ -z "${SSH_AUTH_SOCK}" ]]; then
    SSH_AUTH_SOCK="/run/user/${UID}/ssh-agent.sock"
    if [[ ! -e "${SSH_AUTH_SOCK}" ]]; then
        echo "Starting new ssh-agent..." >&2
        eval $(ssh-agent -s -a "${SSH_AUTH_SOCK}")
    fi
fi

export SSH_AUTH_SOCK
