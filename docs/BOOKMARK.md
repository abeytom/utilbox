# BOOKMARK

The command `bk` is used to manage the following features and is used in conjunction with other wrappers

- Filesystem Path
- Command Alias
- Key Values

### Wrappers

- `bk`
- `op`
- `opf`
- `kv`
- `run`

## Usage

### 1. PATHS

```
# ADD
bk add path [alias] [/path/to/dir] [-f]
bk add path util ~/github/abeytom/utilbox

# LIST
bk list path

# GET
bk get path $alias 
```

#### Alias Usage

```
# cd to '~/github/abeytom/utilbox'
. op util

# open a finder at '~/github/abeytom/utilbox'
opf util
```

### 2. COMMANDS

```
#ADD
bk add cmd $alias $cmd

bk add cmd ssh-vb 'ssh -i /path/to/file abey@192.168.1.131'

bk list cmd

bk get cmd $alias
```

#### Alias Usage

When a command is added an index is added automatically. The command can be executed using the alias or the index

```
run $alias
run ssh-vb
run ssh-vb [arg1] [arg2]

run $aliasIndex
run 0
```

### 3. KEY VALUE

```
bk add kv key value

bk list kv

bk get kv $key
```

#### Alias Usage

```
kv get $key

curl2 -bearer $key https://some.host/some/path/that/takes/basic/auth 
```


