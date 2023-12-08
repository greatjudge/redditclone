FROM golang:1.21

WORKDIR /redditclone

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY ./ ./

RUN go mod tidy
RUN go build -o ./redditclone ./cmd/redditclone/main.go

EXPOSE 8080

RUN chmod +x ./entrypoint.sh
ENTRYPOINT './entrypoint.sh'
