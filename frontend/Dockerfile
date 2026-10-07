# Stage 1: Build React SPA
FROM node:22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY package*.json ./
RUN npm install

# Build static assets
COPY . .
RUN npm run build

# Stage 2: Minimal runtime with Nginx
FROM nginx:alpine

# Remove default nginx website
RUN rm -rf /usr/share/nginx/html/*

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config with reverse proxy
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

