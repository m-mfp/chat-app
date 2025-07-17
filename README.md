# Real-Time Chat App

A real-time chat application showcasing full-stack development with React, Go, PostgreSQL, and WebSockets, containerized with Docker. Messages are sent and received instantly and stored persistently, with planned Auth0 JWT authentication.

## Chat App Setup

### Features
- **Go Backend**: Built with Gin and GORM, handles WebSocket (`/ws`) connections for real-time messaging and HTTP POST (`/api/messages`) for compatibility, plus a `/health` endpoint for monitoring.
- **PostgreSQL Database**: Stores messages persistently using GORM with automatic schema migration for the `Message` model.
- **WebSocket Communication**: Uses `gorilla/websocket` for instant message broadcasting to all connected clients.
- **React Frontend**: A responsive UI for sending and displaying messages in real time, served via Nginx.
- **Dockerized Setup**: Docker Compose orchestrates the React client, Go server, and PostgreSQL database in a single network.
- **Planned Authentication**: Auth0 JWT for user authentication (in progress).

### Running Locally
1. Clone the repo:
   ```bash
   git clone https://github.com/your-username/your-repo.git
   cd your-repo

### Running Locally
1. Clone the repo: `git clone https://github.com/m-mfp/chat-app`
2. Run `docker-compose up --build` to start the server and database
3. Create `client/.env` file `REACT_APP_API_URL=http://server:8000` and `server/.env` file
4. Access the server at `http://localhost:8000/health` and the client at `http://localhost:3000`.
5. Test messages: `curl -X POST http://localhost:8000/api/messages -H "Content-Type: application/json" -d '{"userid": "testuser", "msgid": "123", "text": "Hello, world!"}''`
6. Stop the app `docker-compose down`

### Dockerfile Highlights
- **Server**: Multi-stage Go build for a minimal image, exposing port 8000, with gorilla/websocket and GORM dependencies.
- **Client**: Multi-stage Node.js build with Nginx to serve the React app on port 80, using a custom nginx.conf for optimized static file delivery and SPA routing support.
- **Docker Compose**: Configures networking and loads environment variables from client/.env and server/.env for seamless deployment.
