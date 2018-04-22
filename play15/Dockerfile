FROM openjdk:8-jre-alpine

RUN apk add --update \
    python 

RUN wget https://downloads.typesafe.com/play/1.5.0/play-1.5.0.zip
RUN unzip play-1.5.0.zip

RUN ln -s /play-1.5.0/play /bin/play 
