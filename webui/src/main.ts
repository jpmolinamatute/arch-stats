import {
    checkSessionOpen,
    openSessionFlow,
    closeSessionFlow,
} from './components/sessions/sessions';

const appSection = document.getElementById('app');
const sessionBtn = document.getElementById('session-btn');

async function setupSessionBtn() {
    if (!sessionBtn || !appSection) return;
    const isOpen = await checkSessionOpen();
    if (isOpen) {
        sessionBtn.textContent = 'Close Session';
        sessionBtn.onclick = async () => {
            await closeSessionFlow();
            sessionBtn.textContent = 'Open Session';
            appSection.innerHTML = '';
            setupSessionBtn();
        };
    } else {
        sessionBtn.textContent = 'Open Session';
        sessionBtn.onclick = async () => {
            await openSessionFlow(appSection, setupSessionBtn);
        };
    }
}

function main() {
    setupSessionBtn();
}

// Listen for route changes
window.addEventListener('hashchange', main);
window.addEventListener('DOMContentLoaded', main);
