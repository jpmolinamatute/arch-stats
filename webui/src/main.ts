import { createApp } from 'vue';
import './style.css';
import App from './App.vue';

import { fetchOpenSession } from './state/session';

fetchOpenSession();
const app = createApp(App);
app.mount('#app');

app.config.errorHandler = (err, instance, info) => {
    // report error to tracking services
    console.error('Global error handler:', err);
    console.error('Error info:', info);
    console.error('Component instance:', instance);
};
