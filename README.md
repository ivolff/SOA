### ДЗ5 по сервисам вот оно....

#### Как запускать
* Собираем docker-compose
* запускаем на локалхосте клиент

#### Как это работат

Клиент поднимается на одной системе с очередью задач (можно на разных, но надо хосты поменять, или клиент в докер засунуть, я хотел, но я устал)
Клиент получает 2 url кидает их по RPC(RPC - поверх rabbitmq) на сервер.

Cервер получает запрос с адресами (в отдельной горутине), дает ему ID и кладет в очередь, сам при этом блокируя горутину с запросом, и ожидает когда вернется ответ
Сервер отправляет запрос в rabbitmq, с другой стороны которпого его ждет n воркеров (в моем случае 2, но это легко менять).

Один из воркеров берет запрос и начинает его прасить вертикально поднимая для каждой ветки новую горутину(ограничение по глубине щас стоит 10) (да веротяно поиск в ширину был бы эффективнее, и можно было это распарралелить умнее, но работает)
btw парралельный поиск в глубину в неоктором роде - поиск в ширину так что норм...

Когда нашли (или не нашли) путь, кладем его в обратную очередь.
На сервере достаем ответ из очереди и добавляем по ID в выполненные, откуда его забирает горутина в окторой кртуится RPC возвращает на клиент, и происходит успех

И да для этой задачи можно было не делать сервер, и сразу кидать с клиента в очередь задач, и отправлять назщад по ID отправителя... Но в более сложной системе сервер нужен, да и дз о том чтоб покрутить очередь, так что сделал так


Вообщем вроде все сделал, очередь есть, RPC есть, парсер есть


PS я иногда код написанн ужасно, но я продолжаю преисполняться в GO и переписывать по 10 раз все, каждый раз когда понят что был не прав уже сил нет, не бейте)
