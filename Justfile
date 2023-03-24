all:
    go build

test:
    go run bucket.go "$(mktemp -d)"

release:
    GOOS=linux GOARCH=amd64 go build -o bucket-x86 -v -a
    sha256sum bucket-*

clean:
    rm -f bucket bucket-*
