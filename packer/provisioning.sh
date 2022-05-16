#!/bin/sh

set -exu
arch=$(uname -m)
if [ "$arch" = "x86_64" ]; then
	GOARCH="amd64"
elif [ "$arch" = "aarch64" ]; then
	GOARCH="arm64"
else
	echo "Unknown architecture: $arch"
	exit 1
fi

apt-get update
apt-get install -y docker.io mysql-client-core-8.0 git
update-alternatives --set editor /usr/bin/vim.basic
systemctl enable docker
systemctl start docker
curl -sL https://github.com/docker/compose/releases/download/v2.4.1/docker-compose-linux-${arch} > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

addgroup isucon
adduser isucon --ingroup isucon --ingroup adm --ingroup docker --gecos "" --disabled-password
mkdir /home/isucon/.ssh
chown isucon:isucon /home/isucon/.ssh
chmod 700 /home/isucon/.ssh
echo 'isucon ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/isucon
chmod 440 /etc/sudoers.d/isucon

cd /tmp
tar zxvf /tmp/isucon.tar.gz
cd /tmp/isucon
chown -R isucon:isucon ./
rsync -av ./ /home/isucon/

cd /home/isucon/bench
ln -s bench.$GOARCH bench

cd /home/isucon/webapp
if [ "$arch" = "aarch64" ]; then
	perl -pi -e "s{mysql/mysql-server:8.0.28}{mysql/mysql-server:8.0.28-$arch}" docker-compose.yml
fi
chmod 1777 /home/isucon/webapp/mysql/logs
docker-compose up --build -d
set +xe
while sleep 5; do
	echo .
	mysql -uroot -proot --host 127.0.0.1 --port 13306 -e "select now()" 2> /dev/null
	if [ $? -eq 0 ]; then
		break
	fi
done
curl -si localhost
rm -rf /tmp/isucon*
