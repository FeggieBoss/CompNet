## Задание 1
1. HTTP 1.1 для моего браузера и сервера
2. Accept-Language: en-GB,en-US,ru. Браузер предоставляет: 
    1) Host(имя домена)
    3) Connection(хотим ли мы постоянное соединение)
    4) User-agent(позволяет определить тип браузера)
    5) Accept(какие типы файлов может обработать браузер)
    6) Referer(на какой странице мы были перед тем, как перейти на эту)
    7) Accept-Encoding(алгоритмы сжатия, с которыми браузер готов работать)
3. 192.168.0.107 - мой,	128.119.245.12 - сервера
4. 200OK
5. Last-Modified: Thu, 16 Feb 2023 06:59:02 GMT
6. Content-Length: 128 (128 байт)
## Задание 2
1. Строки IF-MODIFIED-SINCE нет
2. Да, вернул. Line-based text data: ...
3. If-Modified-Since: Thu, 16 Feb 2023 06:59:02 GMT (насколько актуальное значение у меня лежит)
4. 304 Not Modified. Нет, не вернул, так как на сервере нет более актуальное версии данных
## Задание 3
1. Мой браузер отправил один запрос GET. Frame: 28 (28-й)
2. Frame: 30 (30-й)
3. Потребовался один сегмент размера 4927байт
4. В передаваемых данных нет никакой информации заголовка HTTP,связанной с
сегментацией TCP
## Задание 4
1. 3 GET запроса на адреса: 
    1) [Full request URI: http://gaia.cs.umass.edu/wireshark-labs/HTTP-wireshark-file4.html]
    2) [Full request URI: http://gaia.cs.umass.edu/pearson.png]
    3) [Full request URI: http://kurose.cslash.net/8E_cover_small.jpg]
2. Мне кажется, что параллельно, так как сначала идут 2 GET запроса, а потом ответы на них
## Задание 5
1. 401 Unauthorized
2. Authorization: Basic d2lyZXNoYXJrLXN0dWRlbnRzOm5ldHdvcms (Credentials: wireshark-students:network)