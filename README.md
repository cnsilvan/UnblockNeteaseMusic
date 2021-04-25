# UnblockNeteaseMusic

解锁网易云音乐客户端变灰歌曲 (Golang)

[配套LUCI](https://github.com/cnsilvan/luci-app-unblockneteasemusic)

# 特性

* 就是快
* 较为精准的歌曲匹配
* 低内存、高效率
* 暂时支持酷狗、酷我 、咪咕的源
* 支持客户端选择音质（暂时支持酷我、咪咕）
* 支持在搜索页直接显示其他平台搜索结果
* 学习过程中的产物，随缘更新

# 运行

> [release页面](https://github.com/cnsilvan/UnblockNeteaseMusic/releases)中0.2.7及其之后的zip包将默认自带证书，该证书相对比较可靠。  

为了你的安全，还是建议你自己生成证书

```shell
./createCertificate.sh
```

运行程序（由于m=1时 会自动修改hosts生效 所以需要sudo）

```shell
sudo ./UnblockNeteaseMusic
```

### 具体参数说明

```shell
./UnblockNeteaseMusic -h

  -b	force the best music quality
  -c string
    	specify server cert,such as : "server.crt" (default "./server.crt")
  -e	enable replace song url
  -k string
    	specify server cert key ,such as : "server.key" (default "./server.key")
  -l string
    	specify log file ,such as : "/var/log/unblockNeteaseMusic.log"
  -m int
    	specify running mode（1:hosts） ,such as : "1" (default 1)
  -o string
    	specify server source,such as : "kuwo" (default "kuwo")
  -p int
    	specify server port,such as : "80" (default 80)
  -sl int
    	specify the number of songs searched on other platforms(the range is 0 to 3) ,such as : "1"
  -sp int
    	specify server tls port,such as : "443" (default 443)
  -v	display version info

```

# 重要提示

1. 应用通过本机dns获取域名ip，请注意本地hosts文件
2. 受限于歌曲md5的计算时间，耐心等待一会儿再点击下载歌曲吧
3. 网易云APP能用就别升级，不保证新版本可以使用
4. 开启多个源会自动选择最优匹配歌曲，并发的支持使得同时多个源获取并不会增加多少耗时，如果发现一直获取很慢，使用排除法来检查是否某一个源无法使用
6. 0.2.3版本后默认音质将根据客户端选择的音质，启动参数加上`-b`可以设置强制音质优先
7. 0.2.5版本后支持直接在搜索页面显示其他平台搜索结果，如遇错误，请关闭该功能(参数为 `-sl 0`)

### IOS信任证书步骤

1. 安装证书--设置-描述文件-安装
2. 通用-关于本机-证书信任设置-启动完全信任

### 已知

1. windows版本的网易云音乐需要在应用内 设置代理 Http地址为「HttpProxy」下任意地址 端口 80
2. Linux 客户端 (1.2 版本以上需要在终端启动网易云客户端时增加 --ignore-certificate-errors 参数)
3. ios客户端需要信任根证书且运行UnblockNeteaseMusic时 加上 -e 参数
4. android客户端使用咪咕源下载歌曲需要在运行UnblockNeteaseMusic时 加上 -e 参数（其他情况无法使用时，尝试加上 -e 参数）
5. 咪咕源貌似部分宽带无法使用
6. 最新版app采用第三方登录会失败，失败的同学选择手机号登录吧
7. Android7.0后默认不信任用户证书，解决办法自行谷歌
8. 咪咕已无法使用，需要登录，没空写了。且用且珍惜。

# 感谢

[NodeJs版本](https://github.com/nondanee/UnblockNeteaseMusic)以及为它贡献的所有coder

# 声明
该项目只能用作学习，请自行开通会员以支持平台购买更多的版权
