FROM golang:1.8

WORKDIR /go/src/simpl/static
COPY . .

RUN go get ./
RUN go build
RUN ./static
# CMD ["./app > app.txt"]