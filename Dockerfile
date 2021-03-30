## We specify the base image we need for our
## go application
<<<<<<< HEAD

# FROM golang:1.12.0-alpine3.9
FROM golang:1.16.2-alpine3.13

=======
FROM golang:1.16.2
>>>>>>> 36f2bc11759c64cf5e843f828c1f328e4772b73b
## We create an /app directory within our
## image that will hold our application source
## files
RUN mkdir /app
## We copy everything in the root directory
## into our /app directory
ADD . /app
## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /app

## Install Dependency
RUN go get -u github.com/gin-gonic/gin

## we run go build to compile the binary
## executable of our Go program
RUN go build -o main .
## Our start command which kicks off
## our newly created binary executable
CMD ["/app/main"]
LABEL Name=kademliabookmarks Version=0.0.1
EXPOSE 1053/udp 8080
