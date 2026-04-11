# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json ./
RUN npm install --frozen-lockfile || npm install
COPY . .
RUN npm run build

# Stage 2: Serve via nginx
FROM nginx:1.27-alpine
RUN apk add --no-cache openssl && mkdir -p /etc/nginx/ssl && \
	openssl req -x509 -nodes -newkey rsa:2048 \
	-keyout /etc/nginx/ssl/fitcommerce.key \
	-out /etc/nginx/ssl/fitcommerce.crt \
	-days 365 \
	-subj "/CN=localhost"
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000
EXPOSE 3443
CMD ["nginx", "-g", "daemon off;"]
