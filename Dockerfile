FROM alpine:3.7
MAINTAINER Jeremy Edwards <jeremyje@gmail.com>
COPY gowebserver gowebserver

EXPOSE 80
CMD ["gowebserver"]