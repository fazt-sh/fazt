import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { router } from './router.js'
import App from './App.js'
import './styles/admin.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
router.isReady().then(() => app.mount('#app'))
