## seu-wlan V2

seu-wlan 是帮助自动登录东南大学校园网并防止掉线的小工具

### Requirements

* Python3.6+
* requests

### Installation

```sh
pip3 install requests --user
git clone https://github.com/higuoxing/seu-wlan.git
```

### Usage
```
usage: seu-wlan -u username -p password [-t seconds]
```

### Arguments
| Options                   | Usage                                                             |
| ------------------------- | ----------------------------------------------------------------- |
| -u                        | 一卡通号码                                                          |
| -p                        | 统一认证密码                                                        |
| -t                        | 客户端超时时间                                                      |

### Screenshots
![](./.screenshot/seu-wlan-screenshot.jpg)
