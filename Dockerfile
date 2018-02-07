FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get github.com/bitly/go-simplejson github.com/gorilla/mux github.com/tredoe/osutil/user/crypt/sha512_crypt
RUN go build -o main .
CMD ["/app/main"]
