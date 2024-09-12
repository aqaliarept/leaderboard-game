FROM golang:1.22.1-alpine as build

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
	
WORKDIR /
COPY ./src .

RUN go mod download
RUN go build -o app

FROM gcr.io/distroless/static:nonroot

# FROM alpine:latest
WORKDIR /
COPY --from=build /app .
EXPOSE 8000

# USER 65532:65532
ENTRYPOINT ["/app"] 
