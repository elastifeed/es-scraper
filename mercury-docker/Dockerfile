FROM alpine

RUN apk add --update \
    nodejs \
    npm \
    git

COPY main.js .
COPY package.json .
RUN npm install jquery
RUN npm install @postlight/mercury-parser
RUN npm install express

ENTRYPOINT ["/usr/bin/node", "./main.js"]

EXPOSE 8080
