import { reactive } from 'vue';

export const uiManagerStore = reactive({
    currentView: '',
    setView(viewName: string) {
        uiManagerStore.currentView = viewName;
    },
    clearView() {
        uiManagerStore.currentView = '';
    },
});
