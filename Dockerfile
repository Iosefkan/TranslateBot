FROM golang:1.23.4

RUN apt update && apt install -y libasound2-dev
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o translatebot main.go client.go

CMD ["./translatebot"]