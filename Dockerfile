FROM golang:1.19.5-alpine

#Builds main bot
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /app/bot

#Build the image magick module for later
WORKDIR /scripts
COPY 'dockerScripts/*' ./
#RUN './buildImageGen.sh'

#Builds the image gen component
#WORKDIR /app/genImage
#COPY genImage/* ./
#RUN go mod download
#RUN go build -o /app/genImage/imageGen

CMD ["/app/bot"]