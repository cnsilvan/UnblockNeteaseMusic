# UnblockNeteaseMusic
解锁网易云音乐客户端变灰歌曲 (Golang)

# 必读
解决nodejs版本获取灰色歌曲过慢且路由器需有支持VFP的CPU等问题（我的路由器是K3啊，悲剧的K3啊）

目前速度已经大幅度超越nodejs版本了，功能基本一致，有空的话我会继续写下去的！

边学go边写的项目，多点谅解，谢谢！

# 运行
先为自己生成证书（windows需要自己下载openssl）
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
```

重要提示：

本应用获取music.163.com的IP是通过本机直接查询，非nodejs版本请求music.httpdns.c.163.com获取

# 感谢
[NodeJs版本](https://github.com/nondanee/UnblockNeteaseMusic)以及为该项目付出的所有人
