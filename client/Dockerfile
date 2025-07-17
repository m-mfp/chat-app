# Build stage
FROM node:18-alpine AS build

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json .env ./
RUN npm install

# Copy and build React app
COPY . .
RUN npm run build

# Serve stage
FROM nginx:alpine

# Copy built React app to Nginx
COPY --from=build /app/build /usr/share/nginx/html

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]