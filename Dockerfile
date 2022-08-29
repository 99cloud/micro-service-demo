FROM golang:1 as build

WORKDIR /src

COPY . .

RUN make build

FROM alpine:3 as prod

WORKDIR /app

COPY --from=build /src/dist/* .

COPY --from=build /src/html html

