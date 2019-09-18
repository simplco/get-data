FROM golang:1.8

WORKDIR /go/src/simpl/static
COPY . .

RUN go get ./
RUN go build static
RUN ./static > logfile 
# CMD ["./app > app.txt"]