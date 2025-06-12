import { sessionStore } from './state/session_store';
import { openSessionFlow, closeSessionFlow } from './components/sessions_new/sessions_new';
import { renderShotsHistory } from './components/shots_history/shots_history';

const appSection = document.getElementById('app')!;
const sessionBtn = document.getElementById('session-btn')!;

// Responsible for rendering the right component into #app depending on session state
function renderApp(sessionId: string | null) {
    // Clear current content
    appSection.innerHTML = '';

    if (sessionId) {
        // If session is open, show shots history
        renderShotsHistory(appSection, sessionId);
    } else {
        // If session is not open, show the open-session form
        openSessionFlow(appSection, async () => {
            // After successfully opening a session, refresh global session state
            await sessionStore.refresh();
        });
    }
}

// Responsible for updating the session button
function updateSessionBtn(sessionId: string | null) {
    if (sessionId) {
        sessionBtn.textContent = 'Close Session';
        sessionBtn.onclick = async () => {
            await closeSessionFlow();
            await sessionStore.refresh();
        };
    } else {
        sessionBtn.textContent = 'Open Session';
        sessionBtn.onclick = () => {
            renderApp(null);
        };
    }
}

// Reactively rerender on sessionId change
function setupReactivity() {
    sessionStore.subscribe((sessionId) => {
        updateSessionBtn(sessionId);
        renderApp(sessionId);
    });
}

// Main entry point
function main() {
    sessionStore.refresh();
    setupReactivity();
}

// Listen for route changes (if you want, you can keep this for future routing support)
window.addEventListener('hashchange', main);
window.addEventListener('DOMContentLoaded', main);
