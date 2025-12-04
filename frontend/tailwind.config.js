/**
 * Tailwind CSS configuration for Arch-Stats Frontend
 *
 * World Archery (WA) brand palette (see documentation/PRD/frontend/00_UI_requirements.md):
 *  - wa-reflex  (#00209F) Primary backgrounds, headers
 *  - wa-pink    (#E5239D) Accents / secondary buttons
 *  - wa-yellow  (#FFC726) Warnings / emphasis
 *  - wa-green   (#12AD2B) Success / positive states
 *  - wa-red     (#F42A41) Errors / destructive / negative states
 *  - wa-sky     (#00B5E6) Informational highlights
 *  - wa-black   (#000000) Text / borders / dark surfaces
 *
 * In addition to the raw WA colors (namespaced under `wa`), a small set of
 * semantic aliases is provided (primary, accent, success, warning, danger,
 * info, neutral) to encourage intent‑driven styling in components.
 */

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,ts,tsx,js,jsx}'],
  theme: {
    extend: {
      screens: {
        p360: '360px',
        p375: '375px',
        p390: '390px',
        p414: '414px',
      },
      colors: {
        wa: {
          reflex: '#00209F',
          pink: '#E5239D',
          yellow: '#FFC726',
          green: '#12AD2B',
          red: '#F42A41',
          sky: '#00B5E6',
          black: '#000000',
        },
        // Semantic (mirrors WA usage guidance)
        primary: '#00209F', // reflex blue
        accent: '#E5239D', // pink
        success: '#12AD2B', // green
        warning: '#FFC726', // yellow
        danger: '#F42A41', // red
        info: '#00B5E6', // sky
        neutral: '#000000', // black
      },
    },
  },
  plugins: [],
}
