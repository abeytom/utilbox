#! /bin/bash

# Switch the k8s contexts
# ~/.kube/config.name -> name

if [ -z "$1" ]
  then
    echo "Available contexts are "
    cd ~/.kube/ || exit 1
    for name in *; do
        if echo $name | grep -F "config." > /dev/null; then
            extension="${name##*.}"
            echo "  $extension"
        fi
    done
    exit 1
fi


file="$HOME/.kube/config.$1"
if test -f "$file"; then
    cat "$file" > "$HOME/.kube/config"
    chmod 600 "$HOME/.kube/config"
else
  echo "The file doesnt exist [$file]"
  exit 1
fi