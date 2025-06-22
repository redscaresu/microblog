# Microblog

This code powers the ashouri.xyz blog running on scaleway as servlerless containers.

It uses a container which runs the blog application including its cache.

It writes to a serverless postgres instance for persistency.

To run locally simply follow the docker compose commands below.

### Run containers locally

`docker compose up --build`

### Stop containers

`docker compose down -v`

### Run binary locally

`source .env.dev`