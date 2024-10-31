FROM golang:1.21

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Making sure that netcat is installed for health check later on
RUN apt-get update && apt-get install -y netcat-traditional

COPY . .
RUN go build -v -o /usr/local/bin/fts ./main.go

CMD ["fts"]
