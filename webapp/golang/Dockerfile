FROM golang:1.18.1-bullseye
RUN apt-get update && apt-get -y upgrade
RUN mkdir -p /home/isucon
COPY . /home/isucon/webapp
WORKDIR /home/isucon/webapp
RUN go build -o isucon ./... 
CMD ["./isucon"]
