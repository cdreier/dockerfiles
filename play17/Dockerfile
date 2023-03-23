FROM openjdk:11-slim

RUN apt update
RUN apt install -y python3 
RUN apt install -y wget 
RUN apt install -y unzip

RUN wget https://github.com/playframework/play1/releases/download/1.7.1/play-1.7.1.zip
RUN unzip play-1.7.1.zip

RUN chmod +x /play-1.7.1/play

RUN ln -s /play-1.7.1/play /bin/play 
