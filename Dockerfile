FROM semior/baseimage:latest
LABEL maintainer="Semior <ura2178@gmail.com>"

WORKDIR /srv

ENV GOFLAGS="-mod=vendor"

COPY . /srv

RUN go build -mod=vendor -o /go/build/app /srv/app

COPY ./scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 2345

CMD ["/entrypoint.sh"]