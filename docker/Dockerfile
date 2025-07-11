FROM alpine AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
# Ничего лишнего
ARG APP_PORT=8087
ENV APP_PORT=$APP_PORT
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./cert /cert 
COPY ./bin  /bin
EXPOSE $APP_PORT
USER 65535:65535
ENTRYPOINT [ "/bin/app" ]

