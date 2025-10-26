FROM golang:1.25-alpine AS builder

RUN go version
ENV GOPATH=/

COPY ./ ./

# install psql
RUN apk update
RUN apk add postgresql-client

# make wait-for-postgres.sh executable
RUN chmod +x wait-for-postgres.sh

# build go app
# RUN go mod download
RUN go build -o url-shortner cmd/url-shortner/main.go

# #EXPOSE the port
EXPOSE 8082

# Run the executable
CMD ["./url-shortner"] 