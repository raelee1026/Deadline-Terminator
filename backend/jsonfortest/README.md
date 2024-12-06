# JSON 格式

* 抓完要整理成 GET
* deadline: 符合 RFC3339 標準

## GET
GET 請求：返回所有任務的 JSON 陣列，格式如下：
[
  {
    "id": 1,
    "title": "Task 1",
    "deadline": "2024-12-06T23:59:59Z",
    "description": "This is the first task.",
    "deleted": false
  }
]

## POST
POST 請求：接受一個 JSON 任務物件（不包含 id 和 deleted 字段），格式如下：
{
   "title": "New Task",
   "deadline": "2024-12-08T12:00:00Z",
   "description": "This is a new task."
}