import { createApp } from 'vue'

// Import styles first to ensure proper CSS cascade
import './style.css'

// Import Font Awesome
import '@fortawesome/fontawesome-free/css/all.min.css'

// Import API configuration (must be before app creation)
import './api/axiosConfig'

import App from './App.vue'
import router from './router' // Import the router from the router/index.ts file

// Create and mount the app
const app = createApp(App)
app.use(router)
app.mount('#app')
