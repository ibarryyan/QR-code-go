## QRCode

### 二维码生成

api：

```shell
http://localhost:8080/qrcode/gen
```

请求参数：

|   参数名   |    类型  |   是否为空   |
| ---- | ---- | ---- |
|   file   |   file   |   拼图文件   |
|   name   |   string   |  用户姓名    |
|   tc   |   int   |   拼图耗时   |
|   codeType   |  string    |  可选logo或half    |

响应参数：

```shell
{"code":200,"data":"1725069662530.jpg"}
```