# set terminal title to command currently being run

_tsl=$(tput tsl 2>/dev/null)
_fsl=$(tput fsl 2>/dev/null)
if [[ -n $_tsl && -n $_fsl ]]; then
  # Write some info to terminal title.
  # This is seen when the shell prompts for input.
  function precmd {
    print -Pn "${_tsl}%n@%m:%~ %(1j,%j job%(2j|s|); ,)${_fsl}"
  }
  # Write command and args to terminal title.
  # This is seen while the shell waits for a command to complete.
  function preexec {
    printf '%s%s%s' "$_tsl" "$1" "$_fsl"
  }
fi
