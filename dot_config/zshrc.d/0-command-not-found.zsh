# Command not found handler for Ubuntu
if [[ -x /usr/lib/command-not-found ]] ; then
    function command_not_found_handler() {
        /usr/lib/command-not-found --no-failure-msg -- $1
    }
fi

# Command not found for Arch
if [[ -e /usr/share/doc/pkgfile/command-not-found.zsh ]]; then
    source /usr/share/doc/pkgfile/command-not-found.zsh
fi
