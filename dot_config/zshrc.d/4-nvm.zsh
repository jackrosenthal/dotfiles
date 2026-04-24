# Setup nvm if installed; otherwise point npm at ~/.local.
# (nvm refuses to load when npm_config_prefix is set.)

: "${NVM_DIR:=${HOME}/.nvm}"
if [[ -s "${NVM_DIR}/nvm.sh" ]]; then
    unset npm_config_prefix
    export NVM_DIR
    source "${NVM_DIR}/nvm.sh"
else
    export npm_config_prefix="${HOME}/.local"
fi
