import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

import '@fontsource/saira-condensed/600.css';
import '@fontsource/saira-condensed/700.css';
import '@fontsource/saira/400.css';
import '@fontsource/saira/500.css';
import '@fontsource/saira/600.css';
import '@fontsource/chivo-mono/500.css';
import '@fontsource/chivo-mono/600.css';
import './styles/main.css';

createApp(App).use(createPinia()).use(router).mount('#app');
