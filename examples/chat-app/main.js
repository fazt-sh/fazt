// Chat App - Serverless handler for POST messages
// This broadcasts messages via socket and handles API requests

if (req.method === 'POST' && req.body) {
    // Broadcast the message to all WebSocket clients
    socket.broadcast(req.body);
    res.json({
        success: true,
        clients: socket.clients(),
        message: 'Message broadcasted'
    });
} else if (req.path === '/api/status') {
    // Status endpoint
    res.json({
        online: true,
        clients: socket.clients()
    });
} else {
    // For GET requests, we let the static file server handle index.html
    // This file only handles API routes
    res.status(404);
    res.json({ error: 'Use /ws for WebSocket or POST to broadcast' });
}
