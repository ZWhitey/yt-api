FROM node:12
MAINTAINER Whitey

WORKDIR /usr/src/app

COPY package*.json ./

RUN npm ci

COPY index.js .

EXPOSE 3000

CMD ["node", "index.js"]