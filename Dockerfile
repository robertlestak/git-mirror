FROM golang:1.19 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/git-mirror cmd/git-mirror/*.go
RUN go build -o /bin/git-mirrord cmd/git-mirrord/*.go

FROM debian:trixie-slim as app

RUN apt-get update && apt-get install -y git openssh-client ca-certificates

COPY --from=builder /bin/git-mirror /bin/git-mirror
COPY --from=builder /bin/git-mirrord /bin/git-mirrord

ENTRYPOINT ["/bin/git-mirrord"]