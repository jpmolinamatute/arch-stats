import { createScopedStyle } from '../utils/scopedstyle';
import type { components } from '../types/types.generated';
export function LandingPage(): HTMLElement {
    // Elements
    const container = document.createElement('div');
    container.className = 'landing-page';

    // Scoped CSS
    createScopedStyle(
        container,
        `
        .landing-page {
            padding: 2rem;
        }
        .landing-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
        }
        .session-status {
            font-weight: bold;
            font-size: 1.2rem;
            color: #008a45;
        }
        .shots-table {
            border-collapse: collapse;
            width: 100%;
            background: #fff;
            border-radius: 10px;
            box-shadow: 0 2px 8px #0001;
            overflow: hidden;
        }
        .shots-table th, .shots-table td {
            padding: 0.7rem 1rem;
            border-bottom: 1px solid #e0e0e0;
        }
        .shots-table th {
            background: #f5f7fa;
            font-weight: 600;
            text-align: left;
        }
        .shots-table tr:last-child td {
            border-bottom: none;
        }
        .landing-actions button {
            margin-left: 1rem;
            background: #2a2d3a;
            color: #fff;
            border: none;
            border-radius: 7px;
            padding: 0.5rem 1.5rem;
            font-weight: 600;
            font-size: 1rem;
            cursor: pointer;
            transition: background 0.2s;
        }
        .landing-actions button:hover {
            background: #4a90e2;
        }
        .no-session {
            text-align: center;
            color: #888;
            margin-top: 2rem;
            font-size: 1.3rem;
        }
    `,
    );

    // UI Elements
    const header = document.createElement('div');
    header.className = 'landing-header';

    const status = document.createElement('span');
    status.className = 'session-status';
    status.innerText = 'Loading session...';

    const actions = document.createElement('div');
    actions.className = 'landing-actions';
    // Route navigation (you may need to trigger your router)
    actions.innerHTML = `
        <button id="btn-register-arrows">Register Arrows</button>
        <button id="btn-manage-session">Open/Close Session</button>
    `;

    header.appendChild(status);
    header.appendChild(actions);
    container.appendChild(header);

    // Shots Table or No Session Message
    const tableWrapper = document.createElement('div');
    container.appendChild(tableWrapper);

    // Data State
    let sessionId: string | null = null;
    type Shot = components['schemas']['ShotsRead'];
    let shots: Shot[] = [];

    function renderShotsTable() {
        tableWrapper.innerHTML = '';
        if (!shots.length) {
            tableWrapper.innerHTML = `<div class="no-session">No shots recorded in this session yet.</div>`;
            return;
        }

        const table = document.createElement('table');
        table.className = 'shots-table';
        table.innerHTML = `
            <thead>
                <tr>
                    <th>#</th>
                    <th>Arrow ID</th>
                    <th>Engage Time</th>
                    <th>Disengage Time</th>
                    <th>Landing Time</th>
                    <th>X</th>
                    <th>Y</th>
                </tr>
            </thead>
            <tbody>
                ${shots
                    .map(
                        (shot, idx) => `
                    <tr>
                        <td>${idx + 1}</td>
                        <td>${shot.arrow_id}</td>
                        <td>${
                            shot.arrow_engage_time
                                ? new Date(shot.arrow_engage_time).toLocaleTimeString()
                                : '-'
                        }</td>
                        <td>${
                            shot.arrow_disengage_time
                                ? new Date(shot.arrow_disengage_time).toLocaleTimeString()
                                : '-'
                        }</td>
                        <td>${
                            shot.arrow_landing_time
                                ? new Date(shot.arrow_landing_time).toLocaleTimeString()
                                : '-'
                        }</td>
                        <td>${shot.x_coordinate ?? '-'}</td>
                        <td>${shot.y_coordinate ?? '-'}</td>
                    </tr>
                `,
                    )
                    .join('')}
            </tbody>
        `;
        tableWrapper.appendChild(table);
    }

    function renderNoSession() {
        status.innerText = 'No session open';
        tableWrapper.innerHTML = `<div class="no-session">No session open. Please open a session to start recording shots.</div>`;
    }

    // Navigation handlers (update as needed for your router)
    actions.querySelector('#btn-register-arrows')?.addEventListener('click', () => {
        window.location.hash = '#/arrows';
    });
    actions.querySelector('#btn-manage-session')?.addEventListener('click', () => {
        window.location.hash = '#/session';
    });

    // Load session and shots
    async function loadSessionAndShots() {
        status.innerText = 'Loading session...';
        // 1. Get open session
        const resp = await fetch('/api/v0/session/open');
        const json = await resp.json();
        if (json.data && json.data.id) {
            sessionId = json.data.id;
            status.innerText = 'Session: OPEN';
            // 2. Get shots for session
            await loadShots();
            listenWebSocket();
        } else {
            sessionId = null;
            renderNoSession();
        }
    }

    async function loadShots() {
        if (!sessionId) return;
        const resp = await fetch(`/api/v0/shot?session_id=${sessionId}`);
        const json = await resp.json();
        shots = Array.isArray(json.data) ? json.data : [];
        renderShotsTable();
    }

    // Live updates
    let ws: WebSocket | null = null;
    function listenWebSocket() {
        if (!sessionId) return;
        // Close any existing socket
        if (ws) {
            ws.close();
        }
        ws = new WebSocket(`ws://${window.location.host}/api/v0/ws/shot`);
        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                // Filter by session_id if needed
                if (data.session_id === sessionId) {
                    shots.push(data);
                    renderShotsTable();
                }
            } catch (e) {
                console.log(e);
            }
        };
        ws.onclose = () => {
            // Optionally, try to reconnect
        };
    }

    // Init
    loadSessionAndShots();

    return container;
}
