// Session Management
// URL-based sessions for data isolation

const WORDS = [
  'cat', 'dog', 'fox', 'owl', 'bee', 'ant', 'elk', 'bat', 'jay', 'hen',
  'red', 'blue', 'gold', 'jade', 'mint', 'rose', 'sage', 'teal', 'cyan', 'lime',
  'apple', 'berry', 'grape', 'lemon', 'mango', 'peach', 'plum', 'pear', 'kiwi',
  'cloud', 'river', 'stone', 'leaf', 'wave', 'star', 'moon', 'sun', 'sky', 'snow',
  'swift', 'bold', 'calm', 'keen', 'wild', 'free', 'warm', 'cool', 'soft', 'fast'
]

function generateSessionId() {
  const pick = () => WORDS[Math.floor(Math.random() * WORDS.length)]
  return `${pick()}-${pick()}-${pick()}`
}

// Get or create session from URL
export function getSession() {
  const params = new URLSearchParams(location.search)
  let id = params.get('s')

  if (!id) {
    id = generateSessionId()
    const newUrl = new URL(location.href)
    newUrl.searchParams.set('s', id)
    history.replaceState(null, '', newUrl.toString())
  }

  return id
}

// Get shareable URL with session
export function getSessionUrl() {
  return `${location.origin}${location.pathname}?s=${getSession()}`
}

// Generate new session (reset)
export function generateNewSession() {
  const id = generateSessionId()
  const newUrl = new URL(location.href)
  newUrl.searchParams.set('s', id)
  history.replaceState(null, '', newUrl.toString())
  return id
}
