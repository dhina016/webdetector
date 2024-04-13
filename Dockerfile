# Use an official Go runtime as a parent image
FROM golang:1.16-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace.
COPY . /app

# Build the Go app
RUN go mod download
RUN go build -o webdetector

ENTRYPOINT ["webdetector"]
