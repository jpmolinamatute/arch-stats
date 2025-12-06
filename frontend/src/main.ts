import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)

// Register router
app.use(router)

app.mount('#app')

app.config.errorHandler = (err, instance, info) => {
    // report error to tracking services
    console.error('Global error handler:', err)
    console.error('Error info:', info)
    console.error('Component instance:', instance)
}
