FROM php:7.2-apache

RUN apt-get update
RUN apt-get install -y \
  libzip-dev \
  libfreetype6-dev \
  libjpeg62-turbo-dev \
  libpng-dev 
RUN docker-php-ext-install zip \
  && docker-php-ext-install -j$(nproc) iconv \
  && docker-php-ext-configure gd --with-freetype-dir=/usr/include/ --with-jpeg-dir=/usr/include/ \
  && docker-php-ext-install -j$(nproc) gd

RUN docker-php-ext-install mysqli pdo pdo_mysql \
  && docker-php-ext-enable pdo_mysql

COPY  apache/apache2.conf /etc/apache2/apache2.conf