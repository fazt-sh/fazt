// Serverless function - runs on fazt server via Goja
// Access: GET /api/hello

function handler(req) {
  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      message: "Hello from {{.Name}}",
      time: Date.now()
    })
  }
}
