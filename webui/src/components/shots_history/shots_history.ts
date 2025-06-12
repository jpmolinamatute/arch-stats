import { loadComponentAssets } from '../component_loader';
import type { components } from '../../types/types.generated';

type Shot = components['schemas']['ShotsRead'];

async function fetchShotsForSession(sessionId: string): Promise<Shot[]> {
    const response = await fetch(`/api/v0/shot?session_id=${sessionId}`);
    if (!response.ok) throw new Error(`HTTP error: ${response.status}`);
    const json = await response.json();
    return Array.isArray(json.data) ? json.data : [];
}
export async function renderShotsHistory(container: HTMLElement, sessionId: string) {
    const root = await loadComponentAssets(container, 'shots_history');
    const status = root.querySelector('.shots-history-status') as HTMLDivElement;
    const table = root.querySelector('.shots-history-table') as HTMLTableElement;
    const tbody = table.querySelector('tbody') as HTMLTableSectionElement;

    status.textContent = 'Loading shots...';
    table.style.display = 'none';

    try {
        const shots = await fetchShotsForSession(sessionId);

        if (shots.length === 0) {
            status.textContent = 'No shots found for this session.';
            return;
        }

        tbody.innerHTML = shots
            .map(
                (shot) => `
            <tr>
                <td>${shot.id}</td>
                <td>${shot.arrow_id}</td>
                <td>${shot.arrow_engage_time}</td>
                <td>${shot.arrow_disengage_time}</td>
                <td>${shot.arrow_landing_time ?? '-'}</td>
                <td>${shot.x_coordinate ?? '-'}</td>
                <td>${shot.y_coordinate ?? '-'}</td>
            </tr>
        `,
            )
            .join('');
        status.textContent = '';
        table.style.display = '';
    } catch (err) {
        status.textContent = `Error: ${(err as Error).message}`;
        table.style.display = 'none';
    }
}
