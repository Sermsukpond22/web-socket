# E2E Test Suite for Real-time Chat Application

Automated end-to-end (E2E) testing framework for the Real-time Chat Application.

## Test Tier Architecture

- **Tier 1: Feature Coverage (Authentication & Profile)**
  - Register User A (`user_a@test.com`) and User B (`user_b@test.com`)
  - Login as User A and User B to obtain JWT tokens
  - Verify `GET /api/auth/me` with Bearer token authentication
  - Verify missing token rejection (401)

- **Tier 2: Boundary & Friend System**
  - Send friend request from User A to User B (`POST /api/friends/request`)
  - Fetch pending requests for User B (`GET /api/friends/pending`)
  - Accept friend request by User B (`POST /api/friends/accept`)
  - Verify mutual friendship in `GET /api/friends`
  - Restrict chat messaging to non-friend users (User C)

- **Tier 3: Real-time WebSocket Messaging & DB Persistence**
  - Establish WebSocket connection for User A: `ws://localhost:3000/ws?token=<jwt_a>`
  - Establish WebSocket connection for User B: `ws://localhost:3000/ws?token=<jwt_b>`
  - Dispatch real-time chat payload over WS from A to B
  - Receive real-time payload on User B client
  - Verify message persistence via `GET /api/messages/:friend_id`

- **Tier 4: Scenario & Error Handling**
  - Verify WebSocket connection rejection with missing or invalid JWT tokens
  - Handle duplicate friend request attempts gracefully

## Prerequisites

- Node.js 18+ (Node 22 recommended, native `fetch` & `WebSocket` built-in)
- Backend server running on `http://localhost:3000` (or configured via environment variables)

## Execution Instructions

```bash
# Run using Node directly
node run_e2e.js

# Or run using npm
npm test

# Custom Backend & WebSocket URLs
BASE_URL="http://localhost:3000" WS_URL="ws://localhost:3000" node run_e2e.js
```
