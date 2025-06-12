/**
 * Loads HTML and CSS for a component into the given container, with internal caching.
 *
 * @param container The element to render the HTML into.
 * @param componentName The folder and base name for .html/.css files.
 * @returns The root element of the loaded HTML (useful for querying sub-nodes)
 */
const htmlCache: Record<string, string> = {};
const cssLoaded: Set<string> = new Set();

export async function loadComponentAssets(
    container: HTMLElement,
    componentName: string,
): Promise<HTMLElement> {
    const base = `/src/components/${componentName}/${componentName}`;

    // Inject CSS if not already present
    if (!cssLoaded.has(componentName)) {
        const cssText = await fetch(`${base}.css`).then((r) => r.text());
        const style = document.createElement('style');
        style.id = `css-${componentName}`;
        style.textContent = cssText;
        document.head.appendChild(style);
        cssLoaded.add(componentName);
    }

    // Fetch and cache HTML if needed
    if (!(componentName in htmlCache)) {
        const html = await fetch(`${base}.html`).then((r) => r.text());
        htmlCache[componentName] = html;
    }

    // Set container inner HTML to cached HTML
    container.innerHTML = htmlCache[componentName];

    // Return the root element of the component
    // (Assumes your .html uses a single root element)
    return container.firstElementChild as HTMLElement;
}
