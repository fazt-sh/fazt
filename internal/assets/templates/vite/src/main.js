// Main entry point for {{.Name}}
// Works with AND without Vite

const app = document.getElementById('app')

app.innerHTML = `
  <div class="flex items-center justify-center min-h-screen">
    <div class="text-center">
      <h1 class="text-4xl font-bold text-gray-800 mb-4">{{.Name}}</h1>
      <p class="text-gray-600 mb-8">Your Vite-powered app is ready.</p>
      <button id="api-test" class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
        Test API
      </button>
      <p id="api-result" class="mt-4 text-gray-500"></p>
    </div>
  </div>
`

document.getElementById('api-test').addEventListener('click', async () => {
  const result = document.getElementById('api-result')
  try {
    const res = await fetch('/api/hello')
    const data = await res.json()
    result.textContent = JSON.stringify(data, null, 2)
    result.className = 'mt-4 text-green-600 font-mono text-sm'
  } catch (err) {
    result.textContent = 'Error: ' + err.message
    result.className = 'mt-4 text-red-500'
  }
})

console.log('{{.Name}} loaded')
