FROM library/golang:1.6

RUN mkdir -p /go/src/github.com/alderanalytics/statler

COPY . /go/src/github.com/alderanalytics/statler
WORKDIR /go/src/github.com/alderanalytics/statler
RUN go install -v github.com/alderanalytics/statler

CMD ["go-wrapper", "run"]

EXPOSE 5354
