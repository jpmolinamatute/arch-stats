import { createApp } from 'vue';
import './style.css';
import App from './App.vue';

import { fetchOpenSession } from './state/session';

fetchOpenSession();
createApp(App).mount('#app');
