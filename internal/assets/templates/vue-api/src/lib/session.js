// Session helper - persists session ID in localStorage

const SESSION_KEY = '{{.Name}}_session'

export function getSessionId() {
  let id = localStorage.getItem(SESSION_KEY)
  if (!id) {
    id = crypto.randomUUID()
    localStorage.setItem(SESSION_KEY, id)
  }
  return id
}

export function clearSession() {
  localStorage.removeItem(SESSION_KEY)
}
