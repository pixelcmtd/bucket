all:
    go build

test:
    go run bucket.go "$(mktemp -d)"
