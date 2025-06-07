export function Layout(children: HTMLElement): HTMLElement {
    const container = document.createElement('div');
    container.className = "layout";

    // Navigation bar
    const nav = document.createElement('nav');
    nav.className = "navbar";
    nav.innerHTML = `
        <ul>
            <li><a href="#/" id="nav-home">Shots</a></li>
            <li><a href="#/arrows" id="nav-arrows">Arrows</a></li>
            <li><a href="#/session" id="nav-session">Sessions</a></li>
            <li><a href="#/targets" id="nav-targets">Targets</a></li>
            <li><a href="#/history" id="nav-history">History</a></li>
            <li><a href="#/analysis" id="nav-analysis">Analysis</a></li>
        </ul>
    `;

    // Layout
    container.appendChild(nav);
    container.appendChild(children);

    return container;
}
