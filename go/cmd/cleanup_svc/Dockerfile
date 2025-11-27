FROM golang:1.24.3-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /app

FROM scratch
COPY --from=build /app /app
EXPOSE 55000
CMD ["/app"]
