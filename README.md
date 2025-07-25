# gotest_23.07.25
## Описание работы программы
В корневой папке проекта присутствует Makefile, упрощающий запуск и чистый ребилд контейнера.

Цели makefile:
- all - очистка и генерация swagger; билд и запуск контейнера
- clean-build - очистка и генерация swagger; остановка и отчиска кеша, всех контейнеров; билд и запуск контейнера

Дополнительные цели:
- clear-build - очистка кэша и всех контейнеров
- swagger - очистка и генерация swagger компонентов
- clean-swagger - очистка swagger компонентов.

По умолчанию, подключение к swagger UI можно осуществить по следующему адресу:
``
http://localhost:8080/swagger/index.html#/
``

## Используемые сторонние пакеты
- prettySlog - пакет, делающий вывод логгера более читаемым. Источник: [ссылка на репозиторий](https://github.com/GolangLessons/url-shortener/blob/main/internal/lib/logger/handlers/slogpretty/slogpretty.go)
- middlewares/logger - middleware-handler, логирующий детали поступающих запросов [ссылка на репозиторий](https://github.com/GolangLessons/url-shortener/blob/main/internal/http-server/middleware/logger/logger.go)