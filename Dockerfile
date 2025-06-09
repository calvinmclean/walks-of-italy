FROM golang:1.24-bullseye AS build
RUN mkdir /build
ADD . /build
WORKDIR /build
ENV CGO_ENABLED=1
RUN go build -o walks-of-italy ./cmd/walks-of-italy/main.go

FROM debian:bullseye-slim AS production

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app
WORKDIR /app
COPY --from=build /build/walks-of-italy .
ENTRYPOINT ["/app/walks-of-italy"]
