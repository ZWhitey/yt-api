FROM node:12
LABEL MAINTAINER="Whitey"

WORKDIR /usr/src/app

COPY package*.json ./

RUN npm ci

COPY *.js ./

COPY views ./views

EXPOSE 3000

CMD ["node", "index.js"]