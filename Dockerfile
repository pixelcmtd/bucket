FROM golang:1 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY *.go ./
RUN CGO_ENABLED=0 go build -o app

FROM scratch
COPY --from=build app /
EXPOSE 8022
ENTRYPOINT ["/app"]
