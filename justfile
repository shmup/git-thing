default: (commit "10" "disk")

commit n mode="disk": (setup mode) (run n mode) (stats mode)

# playground path based on mode
_playground mode:
    @if [ "{{mode}}" = "ram" ]; then echo "/dev/shm/git-playground"; else echo "playground"; fi

setup mode:
    #!/usr/bin/env bash
    pg=$(just _playground {{mode}})
    rm -rf "$pg"
    mkdir -p "$pg"
    cp -r a b c d e f g h "$pg"/
    git init -q "$pg"
    git -C "$pg" add .
    git -C "$pg" commit -qm "init"

run count mode:
    #!/usr/bin/env bash
    pg=$(just _playground {{mode}})
    echo "using: $pg"
    go run main.go "$pg" {{count}}

stats mode:
    #!/usr/bin/env bash
    pg=$(just _playground {{mode}})
    echo "commits: $(git -C "$pg" rev-list --count HEAD)"
    du -sh "$pg"/.git | cut -f1 | xargs -I{} echo "repo size: {}"

clean mode="disk":
    #!/usr/bin/env bash
    pg=$(just _playground {{mode}})
    rm -rf "$pg"
