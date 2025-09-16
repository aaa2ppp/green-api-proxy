#!/bin/sh

# Скрипт для объединения исходных файлов проекта
# Использование: ./merge_code.sh <директория1> <директория2> ...

for dir in "$@"; do
    # Ищем файлы с нужными расширениями, исключая тестовые миграции
    find "$dir" -type f \
        ! -path './migrations/test/*' \
        \( -name '*.go' -o -name '*.sql' -o -name '*.js' -o -name '*.sh' -o -name '*.md' \
        -o -name 'Dockerfile*' -o -name '*.y*ml' -o -name 'Makefile*' -o -name '*.example' \) \
    | while read -r f; do
        # Убираем ./ в начале пути
        f="${f#./}"
        
        # Выводим содержимое с заголовком
        echo "=== $f ==="
        echo
        cat "$f"
        echo
    done
done
