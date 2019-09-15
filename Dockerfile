FROM golang:1.8

WORKDIR /go/src/simpl
COPY . .

RUN go get ./
RUN go build simpl
RUN ./simpl > simpl.log 
# CMD ["./app > app.txt"]