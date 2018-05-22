FROM alpine:3.7

LABEL maintainer="Phil J. Łaszkowicz <phil@fillip.pro>"
LABEL version="0.0"
LABEL description="Authentication service for Omnijar."

ENV ARUSHA_PORT=8880

EXPOSE ${ARUSHA_PORT}

RUN apk update
RUN apk add ca-certificates

ADD .build/gitlab.com/omnijar/arusha /usr/bin/arusha
ENTRYPOINT ["arusha"]

CMD ["host"]
