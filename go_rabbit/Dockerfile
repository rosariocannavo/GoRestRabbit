# Use the official Golang image as the parent image
FROM golang:1.17-alpine

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Install any dependencies required by the application
RUN go mod download

# Build the application inside the container
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
