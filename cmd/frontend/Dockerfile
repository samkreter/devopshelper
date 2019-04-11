FROM node:8

WORKDIR /usr/src/app

COPY ./cmd/frontend/package*.json ./

RUN npm install

COPY ./cmd/frontend/ .

EXPOSE 8080

CMD [ "npm", "start" ]