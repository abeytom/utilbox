# Utilbox
Some generic utils that I use on my mac to make my life easier.
May be not the best approach out there, but this has grown on me over the years. I initially created this back in the days
to use in my Windows XP with batch (and java).

Note: The wildcard(*) expressions might need to enclosed in single quotes to avoid expansion by shell or use `noglob` to avoid expansion 
# Install

## Linux

### Current Dir
```
curl -L https://github.com/abeytom/utilbox/releases/download/v0.2/utilbox-linux-amd64.tar.gz \
 | tar -xvz && export PATH=$PATH:$(pwd)/utilbox
```
### Home Dir
```
curl -L https://github.com/abeytom/utilbox/releases/download/v0.2/utilbox-linux-amd64.tar.gz \
 | tar -xvz -C $HOME && export PATH=$PATH:$HOME/utilbox
```

## OSX

### Current Dir
```
curl -L https://github.com/abeytom/utilbox/releases/download/v0.2/utilbox-osx.tar.gz \
 | tar -xvz && export PATH=$PATH:$(pwd)/utilbox
```
### Home Dir
```
curl -L https://github.com/abeytom/utilbox/releases/download/v0.2/utilbox-osx.tar.gz \
 | tar -xvz -C $HOME && export PATH=$PATH:$HOME/utilbox
```


# Dev Setup
- Clone this repo to say `~/github/abeytom/utilbox`
- Add `~/github/abeytom/utilbox` to the $PATH
- Build `cd go && make`
- Give exec permissions cmd alias files `chmod +x bk csv ft goexec gw kc kcurl op opf gcart`

# Core Utils `bk` bookmark
This is used to bookmark various FileSystem Paths and Command Aliases

### Add Path
```
# add a path
bk add path <alias> /path/to file
bk add path cu ~/github.abeytom/utilbox

# List Paths
bk list paths

# cd to that dir
. op cu

# open a finder
opf cu
```

### Add command
```
# add a command
bk add cmd <alias> <abey@192.168.1.131>
bk add cmd ssh-dell "ssh -i /path/to/file abey@192.168.1.131"

# list commands
bk list cmd

# execute a command
bk exec ssh-dell

# execute a command by index
bk exec 0     # list command will show the index also

# with args
bk exec alias args1 arg2
```

# Kubernetes Utils 
This is just a wrapper around `kubectl`.


```
# export the namespace as env variable one time for this particular window
export k8ns=my-namespace

# all kubectl default commands will work with this like
kc get pods                         # wrapper for kubectl -n my-namespace get pods
kc get svc

# shortcut commands 

kc logs pod-name* [:index]                  # use wildcards in name (:index -> :0, :2 if there are multiple matches)
kc logt pod-name* [:index]                  # start a tail and follow with last 100 lines
kc logs[logt] -1                                # -1 will match the latest pod

kc ssh pod-name* [:index] [-- bash|sh]      # use wildcards in name (:index -> :0, :2 if there are multiple matches
 
``` 

# Gcloud Utils [WIP]

## 1. List Latest versions of all artifacts [gcArt]
##### Prerequisite: Run this to make sure that the gcloud is setup
```
gcloud beta artifacts packages list --repository=maven-repo --location=us-west1 --format=json
```

#### Run the command
```
gcart                              # lists all latest packages
gcart list 'artifact*'             # list a specific version
```
