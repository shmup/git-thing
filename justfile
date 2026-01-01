playground := "playground"

default: (commit "100")

commit n: setup (run n) stats

setup:
    @rm -rf {{playground}}
    @mkdir -p {{playground}}
    @cp -r a b c d e f g h {{playground}}/
    @git init -q {{playground}}
    @git -C {{playground}} add .
    @git -C {{playground}} commit -qm "init"

run count="100":
    go run main.go {{playground}} {{count}}

stats:
    @echo "commits: $(git -C {{playground}} rev-list --count HEAD)"
    @du -sh {{playground}}/.git | cut -f1 | xargs -I{} echo "repo size: {}"

clean:
    rm -rf {{playground}}
