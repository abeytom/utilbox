# Utilbox

Some generic utils that I use on my mac to make my life easier. May be not the best approach out there, but this has
grown on me over the years. I initially created this back in the days to use in my Windows XP with batch (and java).

Note: The wildcard(*) expressions might need to enclosed in single quotes to avoid expansion by shell or use `noglob` to
avoid expansion

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

## TOOLS

### [1. BOOKMARK](docs/BOOKMARK.md)

#### [PATHS](docs/BOOKMARK.md#1-paths)

- bk
- op
- opf

#### [COMMANDS](docs/BOOKMARK.md#2-commands)

- bk
- run

#### [KEY VALUE](docs/BOOKMARK.md#3-key-value)

- bk
- kv

### [2. TRANSFORM](docs/TRANSFORM.md)

#### [CSV](docs/TRANSFORM.md#1-csv)

- csv
- tsv
- command line screen scraping

#### [JSON](docs/TRANSFORM.md#2-json)

- json transform

#### [YML](docs/TRANSFORM.md#3-yaml)

- yaml transform

### [3. KUBECTL WRAPPER](docs/KUBECTL.md)

- kc
- k8ns

### [4. CURL WRAPPER](docs/CURL_WRAPPER.md)

- curl2

# Dev Setup

- Clone this repo to say `~/github/abeytom/utilbox`
- Add `~/github/abeytom/utilbox` to the $PATH
- Build `cd go && make`
- Give exec permissions cmd alias files `chmod +x bk csv ft goexec gw kc kcurl op opf gcart`
