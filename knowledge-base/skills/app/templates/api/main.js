// Session-Scoped CRUD API Template
// Copy this to api/main.js and customize for your app

var ds = fazt.storage.ds
var kv = fazt.storage.kv

// Generate unique ID
function genId() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

// Parse request
var parts = request.path.split('/').filter(Boolean)
var resource = parts[1]  // 'items', 'todos', etc.
var id = parts.length > 2 ? parts[2] : null
var session = request.query.session || (request.body && request.body.session)

// Require session for all requests
if (!session) {
  return respond(400, { error: 'session required' })
}

// CRUD Operations
if (resource === 'items') {
  // LIST: GET /api/items
  if (request.method === 'GET' && !id) {
    var items = ds.find('items', { session: session })
    return respond({ items: items })
  }

  // GET ONE: GET /api/items/:id
  if (request.method === 'GET' && id) {
    var item = ds.findOne('items', { id: id, session: session })
    if (!item) return respond(404, { error: 'Not found' })
    return respond(item)
  }

  // CREATE: POST /api/items
  if (request.method === 'POST') {
    var body = request.body

    // Validate required fields
    if (!body.name) {
      return respond(400, { error: 'name is required' })
    }

    var item = {
      id: genId(),
      session: session,
      name: body.name,
      description: body.description || '',
      created: Date.now(),
      updated: Date.now()
    }

    ds.insert('items', item)
    return respond(201, item)
  }

  // UPDATE: PUT /api/items/:id
  if (request.method === 'PUT' && id) {
    var body = request.body
    var query = { id: id, session: session }

    // Verify ownership
    var existing = ds.findOne('items', query)
    if (!existing) return respond(404, { error: 'Not found' })

    // Build update object (only include provided fields)
    var updates = { updated: Date.now() }
    if (body.name !== undefined) updates.name = body.name
    if (body.description !== undefined) updates.description = body.description

    ds.update('items', query, updates)

    var updated = ds.findOne('items', query)
    return respond(updated)
  }

  // DELETE: DELETE /api/items/:id
  if (request.method === 'DELETE' && id) {
    var query = { id: id, session: session }

    // Verify ownership before delete
    var existing = ds.findOne('items', query)
    if (!existing) return respond(404, { error: 'Not found' })

    var deleted = ds.delete('items', query)
    return respond({ deleted: deleted })
  }
}

// Stats endpoint example
if (resource === 'stats' && request.method === 'GET') {
  // Check cache first
  var cached = kv.get('stats:' + session)
  if (cached) {
    return respond(cached)
  }

  // Compute stats
  var items = ds.find('items', { session: session })
  var stats = {
    total: items.length,
    recent: items.slice(0, 10),
    computed: Date.now()
  }

  // Cache for 5 minutes
  kv.set('stats:' + session, stats, 300000)

  return respond(stats)
}

// 404 for unknown resources
respond(404, { error: 'Not found' })
