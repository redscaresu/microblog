# Build stage
FROM golang:1.24.2-alpine AS build

# Set the working directory inside the container
WORKDIR /src

# Copy the Go module files first to leverage Docker's caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the project files
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 go build -o /bin/blog ./cmd/main.go

# Final stage
FROM scratch
COPY --from=build /bin/blog /bin/blog
COPY --from=build /src/sql/create_tables.sql /sql/create_tables.sql
ENTRYPOINT ["/bin/blog"]