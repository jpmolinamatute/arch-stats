import type { components } from '../types/types.generated';

export async function createTarget(payload: components['schemas']['TargetsCreate']) {
    try {
        const response = await fetch('/api/v0/target', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });
        const json = await response.json();
        if (response.ok && json.data) {
            return json.data;
        } else {
            console.error('Failed to create target:', json.errors);
            return null;
        }
    } catch (err) {
        console.error('Error creating target:', err);
        return null;
    }
}
