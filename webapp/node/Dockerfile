FROM node:16.14.2-bullseye
RUN apt-get update && apt-get -y upgrade
RUN mkdir -p /home/isucon
COPY . /home/isucon/webapp
WORKDIR /home/isucon/webapp
RUN npm install -g npm@8.6.0 && npm install
CMD ["npx", "nodemon"]
