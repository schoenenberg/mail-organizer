# Start by building the application.
FROM golang:buster as build

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian10:nonroot

WORKDIR /home/nonroot
COPY --from=build --chown=nonroot /go/bin/app ./

ENTRYPOINT ["/home/nonroot/app"]
