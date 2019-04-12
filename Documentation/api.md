
## Table of Contents
* [获取服务器终端sessionId](#获取服务器终端sessionId)
* [Shell终端会话](#Shell终端会话)

## 获取服务器终端sessionId
URL: /v1/terminal

Method: POST

Param: 

| Field    | FieldType | Required | comment  |
| -------- | --------- | -------- | -------- |
| ip       | string    | true     | ip       |
| username | string    | true     | username |
| password | string    | true     | password |
| port     | int       | false    | port     |

Result:

| Field | FieldType | desc      | comment |
| ----- | --------- | --------- | ------- |
| id    | string    | sessionId | id      |

[Back to TOC](#table-of-contents)

## Shell终端会话

URL: ws://ip:port/v1/sockjs/websocket?id

Method: GET

Param: 

| Field    | FieldType | Required | comment   |
| -------- | --------- | -------- | --------- |
| id       | string    | true     | sessionId |


[Back to TOC](#table-of-contents)