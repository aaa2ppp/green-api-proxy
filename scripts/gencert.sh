#!/bin/sh

set -e

mkdir -p ./cert
chmod 700 ./cert

# Создаём приватный ключ (4096 бит)
openssl genrsa -out ./cert/key.pem 4096

# Генерируем самоподписанный сертификат (на 365 дней)
case $(uname) in
MINGW64_NT*)
    openssl req -new -x509 -key ./cert/key.pem -out ./cert/cert.pem -days 365 -subj "//CN=${CN:-localhost}"
    ;;
*)
    openssl req -new -x509 -key ./cert/key.pem -out ./cert/cert.pem -days 365 -subj "/CN=${CN:-localhost}"
    ;;
esac
