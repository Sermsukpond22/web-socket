# Chat App Project Rules & Conventions

This file (GEMINI.md) defines the core rules, architecture, and constraints for this project. All future development should adhere to these guidelines.

## 1. Project Architecture (Backend)
- **Module-based Architecture (Feature-based)**: The backend is organized by features (modules) rather than technical types.
- **Directory**: All features are located under ackend/modules/.
- **Existing Modules**: uth, riend, message, user.
- **Entry Point**: The main entry point is ackend/main.go. Do NOT put it inside nested directories like cmd/server/.

## 2. File Naming Convention (Backend)
- Files inside a module must follow the <module_name>.<type>.go naming convention.
- **Example (auth module)**: 
  - uth.controller.go
  - uth.service.go
  - uth.repository.go
  - uth.route.go
  - uth.controller_test.go
- This ensures clarity when searching for files across multiple modules.

## 3. Testing Convention
- Unit tests should be written in <module_name>.<type>_test.go files.
- Tests should pass without circular dependency issues and should properly import the necessary modules.
- Run go test ./... in the ackend/ directory to verify.

## 4. Environment Variables (.env)
- **No Hardcoded URLs**: Never hardcode localhost:3000 or port numbers in the source code.
- **Backend**: Uses godotenv. Variables like PORT and JWT_SECRET must be loaded from ackend/.env.
- **Frontend**: Uses Vite's import.meta.env. Variables like VITE_API_URL and VITE_WS_URL must be loaded from rontend/.env.
- The frontend uses a Proxy in ite.config.js to route /api and /ws to the backend securely.

## 5. WebSocket & Authentication
- WebSocket connections are established at /ws.
- Authentication for WebSockets requires a JWT token, which can be passed either via the ?token= query parameter or the Authorization: Bearer <token> header.
- The wsHub and connection handlers are located in ackend/websocket/. Local interfaces should be used within the WebSocket package to prevent circular dependencies with other modules.
