FROM golang:1.23-alpine AS build

WORKDIR /src/cmd
COPY go.* *.go /src/
COPY /internal/templates /src/templates
COPY cmd/main.go /src/cmd/

RUN CGO_ENABLED=0 go build -o /bin/blog

FROM scratch
COPY --from=build /bin/blog /bin/blog
COPY --from=build /src/templates /bin/
ENTRYPOINT ["/bin/blog"]