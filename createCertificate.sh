basepath=$(cd `dirname $0`; pwd)
echo $basepath
extFile="$basepath/extFile.txt"
serverCrt="$basepath/server.crt"
serverKey="$basepath/server.key"
serverCsr="$basepath/server.csr"
caCrt="$basepath/ca.crt"
caKey="$basepath/ca.key"
# 生成 CA 私钥
openssl genrsa -out "${caKey}" 2048
# 生成 CA 证书
openssl req -x509 -new -nodes -key "${caKey}" -sha256 -days 825 -out "${caCrt}" -subj "/C=CN/CN=UnblockNeteaseMusic Root CA/O=UnblockNeteaseMusic"
# 生成服务器私钥
openssl genrsa -out "${serverKey}" 2048
# 生成证书签发请求
openssl req -new -sha256 -key "${serverKey}" -out "${serverCsr}" -subj "/C=CN/L=Hangzhou/O=NetEase (Hangzhou) Network Co., Ltd/OU=IT Dept./CN=*.music.163.com"
# 使用 CA 签发服务器证书
touch "${extFile}"
echo "basicConstraints=CA:FALSE
keyUsage=digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:music.163.com,DNS:*.music.163.com" >"${extFile}"
openssl x509 -req -extfile "${extFile}" -days 825 -in "${serverCsr}" -CA "${caCrt}" -CAkey "${caKey}" -CAcreateserial -out "${serverCrt}"
rm -f "${extFile}"