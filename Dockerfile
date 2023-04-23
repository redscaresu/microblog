FROM golang:1.20-alpine AS build

WORKDIR /src/cmd
COPY go.* *.go /src/
COPY /templates /src/templates
COPY cmd/main.go /src/cmd/

# RUN CGO_ENABLED=0 go build -o /bin/blog

# FROM scratch
# COPY --from=build /bin/blog /bin/blog
# COPY --from=build /templates /bin/
# ENTRYPOINT ["/bin/blog"]