# Test-task-TrueConf
Test task for TrueConf. Junior-Go.

1. Мой первый коммит "StructuredCode" просто сгруппировал все структуры и функци до main, просто для моего визуального комфорта.

2. Во втором коммите я создала: структуру App, которая будет содержать информацию о хранилище, обработчиках и настройках маршрутов, функцию-конструктор NewApp для App, которая будет инициализировать все необходимые компоненты. Вынесла инициализацию маршрутизатора и хранилища в отдельные методы initRouter и initStore структуры App. Сгруппировала обработчики, относящиеся к пользователям, в отдельную функцию initUserRoutes. Переместила обработчики в методы структуры App для более "чистого" и организованного кода. В main функции создала экземпляр приложения и запустила сервер.

3. В третьем коммите я добавила библиотеку логгирования logrus для более гибкого и настраиваемого логгирования и все, чтобы код корректно работал. 

Что можно было бы еще добавить:
1. Конфигурация.
2. Обработка миграций данных.
3. Тесты для ваших обработчиков и функций.
4. Создайте документацию (но думаю, что она у вас есть для более сложных приложений).
5. Сегментирование кода: Если приложение станет более сложным, вы можете рассмотреть возможность разделения кода на несколько пакетов или слоев. Я помню, что в письме говорилось, что приложение будет расти.
6. И туда же контроль версий API.
