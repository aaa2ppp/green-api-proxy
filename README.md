# Green API Proxy

Минималистичный прокси для работы с GREEN API через браузер без CORS-ошибок.

## Быстрый старт

1. Клонируем и настраиваем:
```bash
git clone https://github.com/yourname/green-api-proxy.git
cd green-api-proxy
cp env.example .env
```

2. Генерируем сертификат (укажите свой домен):
```bash
CN=my-domain.local make cert
```

3. Запускаем:
```bash
make build && make run
```

4. Открываем в браузере:
```
https://localhost:8087/green-api/
```

## Что под капотом

- Автоматически обрабатывает CORS
- Проксирует запросы к `api.green-api.com`
- Самоподписанный SSL (для тестов)

Для продакшена замените сертификат на Let's Encrypt
