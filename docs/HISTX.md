# Extended History `histx`

Saves the `time`, `dir` and `zsh session id` into a separate file addition to the default file

## `~/zshrc` .additions

```
HISTFILE=~/.zsh_history
export HISTSIZE=1000000000
export SAVEHIST=$HISTSIZE
setopt INC_APPEND_HISTORY_TIME
setopt HIST_FIND_NO_DUPS
function zshaddhistory() {
    val=${1%%$'\n'}
    if [[ "$val" == "" ]]; then
        return 1
    fi    
    print -sr -- ${val}
    # now we add to this file with some additional data
    fc -p ~/.zsh_history_detail
    SID=$(stat | awk '{print $7}')
    print -sr -- "^^^$(date '+%Y-%m-%d_%R'),${PWD},${SID},${val}"
    return 1
}
```

## `histx` usage

Usage `histx [options]`

Options

- `-d --dir`       : Show the commands executed in the given dir
- `-p --pwd`       : Show the commands executed in the current dir
- `-c --cmd`     : Filter history by a particular command
- `-s --session` : Filter history by a particular zsh session
- `      --uniq`   : unique commands i.e dont show any repeated commands
- `      --suniq`  : don't show repeated successive commands
- `      --co`  : output the command only    


