# **UI & Look-and-Feel Requirements**

## 1. General Style

* **Framework**: Use **Tailwind CSS** utilities as much as possible.
* **Color Palette**: World Archery brand colors (see table below).
* **Theme**: **Dark mode only** (for now).
* **Font**: **Inter** (Google Fonts) for clean, modern typography.
* **Design Style**: Minimalist, easily adjustable later if branding changes.
* **Corners**: Rounded corners for all cards, panels, buttons, and inputs.

For global CSS use `frontend/src/style.css` file

---

## 2. World Archery Color Palette

| Color Name                    | Tailwind Key | HEX       | Intended Usage                                           |
| ----------------------------- | ------------ | --------- | -------------------------------------------------------- |
| **Reflex Blue** *(Corporate)* | `wa-reflex`  | `#00209F` | Primary background, headers, key UI areas                |
| **Pink**                      | `wa-pink`    | `#E5239D` | Accent highlights, secondary buttons                     |
| **Yellow**                    | `wa-yellow`  | `#FFC726` | Warnings, emphasis text/icons                            |
| **Green**                     | `wa-green`   | `#12AD2B` | Success states, positive boolean indicators              |
| **Red**                       | `wa-red`     | `#F42A41` | Errors, destructive actions, negative boolean indicators |
| **Sky Blue**                  | `wa-sky`     | `#00B5E6` | Informational highlights, secondary accents              |
| **Black**                     | `wa-black`   | `#000000` | Text in dark mode, backgrounds, borders                  |

**Tailwind Config Snippet:**

```js
module.exports = {
  theme: {
    extend: {
      colors: {
        wa: {
          reflex: '#00209F',
          pink: '#E5239D',
          yellow: '#FFC726',
          green: '#12AD2B',
          red: '#F42A41',
          sky: '#00B5E6',
          black: '#000000',
        }
      }
    }
  }
}
```

---

## 3. Layout & Structure

* **Toolbar**:

  * Fixed at the top; always visible on scroll.
  * Text buttons only (no icons for now).
  * Compact spacing (reduced padding/margin).
* **Content Area**:

  * Full-width layout for forms, tables, and cards.
* **Padding & Spacing**:

  * Compact design - minimal whitespace between UI elements to maximize data density.

---

## 4. Component Styling

* **Tables**:

  * Striped rows.
  * No hover highlight.
  * Borders between rows.
* **Forms**:

  * Labels above input fields.
  * All "create" forms are displayed inside the Wizard module, not as standalone forms on pages.
  * Consistent **Next / Finish / Cancel** button placement.
* **Buttons**:

  * Solid filled buttons.
  * Rounded corners.
  * Use `wa-reflex` (primary) and `wa-pink` (secondary) for backgrounds.
* **Panels & Cards**:

  * Directly on page background (no card shadows or separate panels unless necessary).
* **Boolean Indicators**:

  * **True** -> Green circle (`wa-green`)
  * **False** -> Red circle (`wa-red`)

---

## 5. Interactions & Feedback

* Minimal animations; prefer instant changes.
* Use consistent button states (default, hover, disabled).
* No hover-based interactions on tables.

---

## 6. Responsiveness

* **Desktop-first** approach.
* Mobile responsiveness is not a priority in the initial phase.
* Layout and component sizing should still degrade gracefully if viewed on smaller screens.

## 7. Tool bar

* one for the session page
* one for the arrow page
