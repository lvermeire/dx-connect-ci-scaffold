FROM node:22-alpine@sha256:4d64b49e6c891c8fc821007cb1cdc6c0db7773110ac2c34bf2e6960adef62ed3 AS builder
WORKDIR /app

# Cache npm install layer separately from source
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine@sha256:645eda1c2477aaa9b879f73909b9222c6f19798dd45be6706268d82a661c6e6d
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
