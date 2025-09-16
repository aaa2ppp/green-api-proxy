FROM alpine AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
# Сертификаты CA и TLS-сертификат приложения
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./cert/cert /cert

# Бинарник приложения
COPY ./bin/app /bin/app

# Порт по умолчанию
ENV SERVER_ADDR=:8080
EXPOSE 8080

# Безопасность: не-root пользователь
USER 65535:65535

ENTRYPOINT [ "/bin/app" ]
