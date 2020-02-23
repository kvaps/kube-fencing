############################
# STEP 1 build executable binary
############################
FROM golang:1.13-alpine3.11 AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk add --no-cache git

COPY . $GOPATH/src/github.com/kvaps/kube-fencing/
WORKDIR $GOPATH/src/github.com/kvaps/kube-fencing/cmd/switcher

# Fetch dependencies.

# Using go get.
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/fencing-switcher

############################
# STEP 2 build a small image
############################
FROM scratch

# Copy our static executable.
COPY --from=builder /go/bin/fencing-switcher /fencing-switcher

ENTRYPOINT ["/fencing-switcher"]
