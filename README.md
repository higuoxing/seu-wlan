## seu-wlan

seu-wlan 是帮助自动登录东南大学校园网并防止掉线的小工具

### Requirements

Go 1.11

### Installation

```sh
$ go get github.com/vgxbj/seu-wlan
```

### Usage
```
usage: seu-wlan -u username -p password [-i seconds] [-enable-mac-auth] [-disable-tls-verification]
```

### Arguments
| Options                   | Usage                                                             |
| ------------------------- | ----------------------------------------------------------------- |
| -u                        | 一卡通号码                                                          |
| -p                        | 统一认证密码                                                        |
| -i                        | 设置轮询登录间隔，以秒为单位 (int)                                     |
| -c                        | 如不想使用明文密码，可以使用配置文件                                    |
| -workers                  | 启用多个客户端进行请求                                                |
| -timeout                  | 客户端超时时间                                                      |
| -enable-mac-auth          | 允许记住本机 mac 地址                                                |
| -disable-tls-verification | 偶尔会出现学校证书过期的情况，可以使用该选项跳过证书校验 (不推荐)            |

### Configuration
参见 ``config.json``

### Screenshots
![](./.screenshot/seu-wlan-screenshot.jpg)
