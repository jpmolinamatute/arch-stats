import { reactive, watch } from 'vue';
import { openSession } from './session';
export type ViewName = 'ArrowForm' | 'SessionForm' | 'TargetForm' | 'ShotsTable';
export const uiManagerStore = reactive({
    currentView: null as ViewName | null,
    setView(viewName: ViewName) {
        this.currentView = viewName;
    },
    clearView() {
        this.currentView = null;
    },
});

watch(
    () => openSession.is_opened,
    (isOpened) => {
        if (isOpened && !uiManagerStore.currentView) {
            uiManagerStore.currentView = 'ShotsTable';
        } else if (isOpened === false && uiManagerStore.currentView === 'ShotsTable') {
            uiManagerStore.currentView = null;
        }
    },
);
