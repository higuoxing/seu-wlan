seu-wlan
=========

seu-wlan 是帮助自动登录校园网并防止掉线的小工具

Requirements
------------
Go 1.11

Installation
------------
.. code:: sh
  $ go get github.com/Higuoxing/seu-wlan

Usage
-----
.. code:: sh
  usage: seu-wlan -u username -p password [-i seconds] [-m 0|1]

positional arguments:
  -u                      一卡通号码
  -p                      统一认证密码

optional arguments:
  -i                      设置轮询登录间隔，以秒为单位 (int)
  -m                      设置是否允许记住 mac 地址 (1|0)

Screenshots
-----------
.. image:: ./.screenshot/seu-wlan-screenshot.jpg
