# Setup nvm, if installed.

: "${NVM_DIR:=${HOME}/.nvm}"
if [[ -d "${NVM_DIR}" ]]; then
    export NVM_DIR
    source "${NVM_DIR}/nvm.sh"
fi
