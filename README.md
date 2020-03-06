# UnblockNeteaseMusic
解锁网易云音乐客户端变灰歌曲 (Golang)

[配套LUCI](https://github.com/cnsilvan/luci-app-unblockneteasemusic)
# 特性
* 就是快
* 较为精准的歌曲匹配
* 全平台支持
* 低内存、高效率
* 暂时支持酷狗、酷我 、咪咕的源（酷我、咪咕支持无损音乐）
* 学习过程中的产物，随缘更新

# 运行
先为自己生成证书（windows需要自己下载openssl）（为了你的安全，请务必自己生成证书）
```
./createCertificate.sh
```

运行程序（由于m=1时 会自动修改hosts生效 所以需要sudo）
```
sudo ./UnblockNeteaseMusic
```

具体参数说明
```
./UnblockNeteaseMusic -h
    -c string
      	specify server cert,such as : "server.crt" (default "./server.crt")
    -e	replace song url
    -k string
      	specify server cert key ,such as : "server.key" (default "./server.key")
    -m int
      	specify running mode（1:hosts） ,such as : "1" (default 1)
    -o string
      	specify server source,such as : "kuwo:kugou" (default "kuwo:kugou")
    -p int
      	specify server port,such as : "80" (default 80)
    -sp int
      	specify server tls port,such as : "443" (default 443)
    -v	display version info

```

重要提示：

本应用获取music.163.com的IP是通过本机直接查询，非nodejs版本请求music.httpdns.c.163.com获取

已知：
1. windows版本的网易云音乐需要在应用内 设置代理 Http地址为「HttpProxy」下任意地址 端口 80
2. Linux 客户端 (1.2 版本以上需要在终端启动网易云客户端时增加 --ignore-certificate-errors 参数)
3. ios客户端需要信任根证书且运行UnblockNeteaseMusic时 加上 -e 参数
4. 咪咕源貌似部分宽带无法使用
# 感谢
[NodeJs版本](https://github.com/nondanee/UnblockNeteaseMusic)以及为它贡献的所有coder
