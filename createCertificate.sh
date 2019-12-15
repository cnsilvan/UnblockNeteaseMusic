basepath=$(cd `dirname $0`; pwd)
echo $basepath
openssl req -new -nodes -x509 -out $basepath/server.crt -keyout $basepath/server.key -days 3650 -subj "/C=CN/ST=SH/L=SH/O=UnblockNeteaseMusic/OU=UnblockNeteaseMusic Software/CN=127.0.0.1/emailAddress=cnsilvan@gmail.com"