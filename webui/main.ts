import { Layout } from "./components/layout";
import { ShotsPage } from "./pages/shotspage";
import { ArrowsPage } from "./pages/arrowspage";
import { SessionsPage } from "./pages/sessionspage";
import { TargetsPage } from "./pages/targetspage";
import { LandingPage } from "./pages/landingpage";

const routes: Record<string, () => HTMLElement> = {
    "/": LandingPage,
    "/shot": ShotsPage,
    "/arrows": ArrowsPage,
    "/sessions": SessionsPage,
    "/targets": TargetsPage,
    // ...add others
};

function render() {
    const app = document.getElementById('app');
    if (!app) return;
    // Parse hash
    const path = location.hash.replace(/^#/, "") || "/";
    const Page = routes[path] || ShotsPage;
    const pageEl = Page();
    // Clear and render
    app.innerHTML = "";
    app.appendChild(Layout(pageEl));
}

// Listen for route changes
window.addEventListener("hashchange", render);
window.addEventListener("DOMContentLoaded", render);
