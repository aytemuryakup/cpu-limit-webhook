#
# Build go project
#
FROM golang as go-builder

WORKDIR /go/src/mutatingwebhook

COPY . .
RUN go mod init
RUN go get -t .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mutatingwebhook *.go

#
# Runtime container
#
FROM alpine:latest  

RUN mkdir -p /app && \
    addgroup -S app && adduser -S app -G app && \
    chown app:app /app

WORKDIR /app

COPY --from=go-builder /go/src/mutatingwebhook .

USER app

CMD ["./mutatingwebhook"]  

