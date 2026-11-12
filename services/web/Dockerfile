FROM node:24-alpine@sha256:116b800c3754b440ee0aa3ed2f20f5bf3d99a554bfd9adeea09ea8017f6e85f5 AS builder
WORKDIR /app

# Cache npm install layer separately from source
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine@sha256:582c496ccf79d8aa6f8203a79d32aaf7ffd8b13362c60a701a2f9ac64886c93d
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
