// utils/scopedStyle.ts

/**
 * Attach a <style> tag scoped to a parent element.
 * Usage: createScopedStyle(parentEl, `.class { ... }`)
 */
export function createScopedStyle(parent: HTMLElement, css: string) {
    const style = document.createElement('style');
    style.textContent = css;
    parent.appendChild(style);
}
