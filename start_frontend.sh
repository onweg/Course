#!/bin/bash
# Скрипт для запуска frontend клиента
# Использование: ./start_frontend.sh

echo "Запуск frontend сервера на порту 8000..."
cd frontend

# Проверяем наличие Python3
if command -v python3 &> /dev/null; then
    python3 -m http.server 8000
elif command -v python &> /dev/null; then
    python -m SimpleHTTPServer 8000
else
    echo "Python не найден. Установите Python или используйте другой HTTP сервер."
    echo "Альтернатива: npx http-server -p 8000"
fi

