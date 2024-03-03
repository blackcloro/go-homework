# syntax=docker/dockerfile:1

# Use the official Golang image as a parent image
FROM golang:1.22

# Set the working directory in the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the working directory inside the container
COPY . .

# Build the application
RUN go build -o main -ldflags="-s -w" ./cmd/api

# Expose the port that the application listens on
EXPOSE 8080

# Set the entrypoint for the container
ENTRYPOINT [ "./main" ]
