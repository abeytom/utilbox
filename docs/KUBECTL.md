# KUBECTL Wrapper

This is a wrapper around `kubectl` command without having to type the `-n $ns` arg with every command. It also has some
shortcut commands

## USAGE

Set the namespace env variable

```
. k8ns my-namespace 

# OR
export k8ns=my-namespace
```

The `kc` wrapper will access this variable from the env

```
# all kubectl default commands will work with this like
kc get pods                                 # wrapper for kubectl -n my-namespace get pods
kc get svc

# shortcut commands 

kc logs pod-name* [:index]                  # use wildcards in name (:index -> :0, :2 if there are multiple matches)
kc logt pod-name* [:index]                  # start a tail and follow with last 100 lines
kc logs[logt] -1                            # -1 will match the latest pod

kc ssh pod-name* [:index] [-- bash|sh]      # use wildcards in name (:index -> :0, :2 if there are multiple matches
kc ssh -1                                   # -1 will match the latest pod

# INLINE
kc describe pod `kc pod pod-name*`
kc describe pod `kc pod pod-name* :index`
```

## (BASH | ZSH) PROMPT

This prompt will add the k8s context and k8s namespace to the prompt.

### BASH PROMPT

```
get_base_dir_prompt(){
    base1="${PWD##*/}"
    dir1="${PWD%/*}"
    PROMPT_STR="${dir1##*/}/$base1"
    echo $PROMPT_STR
}

get_k8ns_prompt(){
    PROMPT_STR=""
    if [[ ! -z "${k8ns}" ]]; then
        PROMPT_STR="(k8ns:$k8ns) (k8ctx:$(kubectl config current-context))"
    fi
    echo $PROMPT_STR
}

# APPEND TO EXISTING PROMPT
PS1+="\[\e[0;34m\]\$(get_k8ns_prompt)\[\e[0m\] \$ "

# CREATE NEW PROMPT 'user@host parent/dir $k8s'
#export PS1="\[\e[1;32m\]\u@\h\[\e[0m\] \[\e[0;35m\]\$(get_base_dir_prompt)\[\e[0m\] \[\e[0;34m\]\$(get_k8ns_prompt)\[\e[0m\] \$ "

```

### ZSH PROMPT (macOS)

```
get_prompt(){
    base1="${PWD##*/}"
    dir1="${PWD%/*}"
    PROMPT_STR="%F{magenta}${dir1##*/}/$base1%F"

    if [[ ! -z "${k8ns}" ]]; then
        PROMPT_STR="$PROMPT_STR %F{blue} (k8ns: $k8ns) (k8ctx:$(kubectl config current-context))%F"
    fi
    
    echo $PROMPT_STR
}
#https://github.com/git/git/blob/master/contrib/completion/git-prompt.sh
source ~/.zsh/git-prompt.sh

# unstaged(*) staged (+) untracked(%) behind(<) ahead(>) 
GIT_PS1_SHOWDIRTYSTATE=true
GIT_PS1_SHOWUPSTREAM="verbose"
GIT_PS1_SHOWUNTRACKEDFILES=true
# GIT_PS1_SHOWCOLORHINTS=true

setopt PROMPT_SUBST
PROMPT='$(get_prompt)%F{green}$(__git_ps1 " (%s)")%F \$ '
```


