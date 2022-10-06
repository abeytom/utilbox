
## zsh history
```
#### ZSH HISTORY ####

HISTFILE=~/.zsh_history
export HISTSIZE=1000000000
export SAVEHIST=$HISTSIZE
setopt INC_APPEND_HISTORY_TIME
setopt HIST_FIND_NO_DUPS
function zshaddhistory() {
    # This is the default history
    print -sr -- ${1%%$'\n'}
    # now add to this file with some additional data
    fc -p ~/.zsh_history_detail
    print -sr -- "^^^$(date '+%Y-%m-%d_%R'),${PWD},${1%%$'\n'}"
    return 1
}
```
