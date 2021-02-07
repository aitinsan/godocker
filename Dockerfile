# Build executable stage
FROM golang
ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o snippetbox cmd/web/*
ENTRYPOINT /app/snippetbox
# Build final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=0 /app/snippetbox .
RUN mkdir ui
COPY ./ui /app/ui
EXPOSE 4000
ENTRYPOINT ["./snippetbox", "-host=host.docker.internal", "-port=5433"]