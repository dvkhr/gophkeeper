# gophkeeper
A secure password manager written in Go
Описание:
GophKeeper - консольное приложение для безопасного хранения данных. Оно позволяет сохранять пары логин-пароль, данные банковских карт,
произвольные текстовые и бинарные данные. Все данные хранятся на сервере в зашифрованном на стороне клиента виде.
Особенности:
Локальное шифрование с использованием мастер-пароля.
Синхронизация между устройствами.
Поддержка типов данных "loginpass", "card", "text".
gRPC API, JWT-аутентификация.
Refresh-токен с отзывом.
Автоматическое обновление сессии.

Запуск сервера:
./build/gophkeeper-server

Использование клиента:
Регистрация ./build/gophkeeper-client register --login vasia --password "mypass"
Добавление данных: 
./build/gophkeeper-client add --id=note1 --type=text --content="Важная заметка"
./build/gophkeeper-client add --id=gmail --type=loginpass --login=user@gmail.com --password="secure123" --meta "yandex.ru"
./build/gophkeeper-client add --id=card1 --type=card --number="1111 1111 1111 1111" --expiry="12/27" --cvv="123"
Получение данных: ./build/gophkeeper-client get
Удаление данных: ./build/gophkeeper-client delete --id=note1
Вход: ./build/gophkeeper-client login --login vasia --password "mypass"
Выход: ./build/gophkeeper-client logout
Версия: ./build/gophkeeper-client version