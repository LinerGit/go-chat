/**
 * Example WebSocket client for Chat Service
 * 
 * 1. Get access token from Auth Service:
 *    POST /auth/login  → { access_token, refresh_token }
 * 
 * 2. Connect to Chat Service WebSocket with token as query param
 */

const ACCESS_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."; // from auth service
const WS_URL = `ws://localhost:8080/ws/chat?token=${ACCESS_TOKEN}`;

const ws = new WebSocket(WS_URL);

ws.onopen = () => {
  console.log("Connected to chat");

  // Send a message
  ws.send(JSON.stringify({
    content: "Hello everyone!"
  }));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  // msg shape:
  // {
  //   type:         "message",
  //   from_user_id: 123,
  //   username:     "john",
  //   content:      "Hello everyone!",
  //   timestamp:    "2026-01-01T12:00:00Z"
  // }
  console.log(`[${msg.username}]: ${msg.content}`);
};

ws.onerror = (err) => {
  console.error("WebSocket error:", err);
};

ws.onclose = (event) => {
  console.log("Disconnected:", event.code, event.reason);
};
