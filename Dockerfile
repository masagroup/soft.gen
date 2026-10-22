# The image of the generators: soft.gen, every language included, so that
#   docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.gen go -m /gen/model.ecore -o /gen
# runs "soft.gen go -m /gen/model.ecore -o /gen".
ARG GO_VERSION=1.27.3

# The build stage is discarded: only the binary reaches the image.
FROM golang:${GO_VERSION}-alpine AS build
ARG VERSION=v2.0.0-developer-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=${VERSION}" -o /soft.gen ./cmd/soft.gen

# soft.gen is a static binary that only reads and writes files: it runs on an
# empty image, which has no operating system package to be vulnerable.
FROM scratch
COPY --from=build /soft.gen /soft.gen
ENTRYPOINT ["/soft.gen"]
