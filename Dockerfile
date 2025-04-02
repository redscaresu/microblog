FROM golang:1.24.2-alpine AS build

# Copy the entire project into the container
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 go build -o /bin/blog ./cmd/main.go

RUN CGO_ENABLED=0 go build -o /bin/blog

FROM scratch
COPY --from=build /bin/blog /bin/blog
COPY --from=build /src/templates /bin/
ENTRYPOINT ["/bin/blog"]