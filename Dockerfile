FROM golang:1.14.4
RUN go get golang.org/x/tools/cmd/godoc
RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /usr/bin/godoc-service .

ENTRYPOINT [ "godoc-service" ]
