// CRUD API for items using fazt storage
// Access: /api/items and /api/items/:id

var ds = fazt.storage.ds

function handler(req) {
  var parts = req.path.split('/').filter(Boolean)
  var id = parts.length > 2 ? parts[2] : null

  // GET /api/items - list all
  if (req.method === 'GET' && !id) {
    var items = ds.find('items', {})
    return respond({ items: items })
  }

  // GET /api/items/:id - get one
  if (req.method === 'GET' && id) {
    var item = ds.findOne('items', { id: id })
    if (!item) return respond(404, { error: 'Not found' })
    return respond(item)
  }

  // POST /api/items - create
  if (req.method === 'POST') {
    var body = req.body
    if (!body.name) return respond(400, { error: 'name required' })

    var item = {
      id: fazt.uuid(),
      name: body.name,
      created: Date.now()
    }
    ds.insert('items', item)
    return respond(201, item)
  }

  // PUT /api/items/:id - update
  if (req.method === 'PUT' && id) {
    var body = req.body
    var updated = ds.update('items', { id: id }, body)
    if (updated === 0) return respond(404, { error: 'Not found' })
    return respond({ updated: updated })
  }

  // DELETE /api/items/:id - delete
  if (req.method === 'DELETE' && id) {
    var deleted = ds.delete('items', { id: id })
    return respond({ deleted: deleted })
  }

  return respond(405, { error: 'Method not allowed' })
}
