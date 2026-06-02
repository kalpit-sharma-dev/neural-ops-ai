---
name: NeuralOps
colors:
  surface: '#131319'
  surface-dim: '#131319'
  surface-bright: '#393840'
  surface-container-lowest: '#0e0e14'
  surface-container-low: '#1b1b21'
  surface-container: '#1f1f26'
  surface-container-high: '#2a2930'
  surface-container-highest: '#35343b'
  on-surface: '#e4e1ea'
  on-surface-variant: '#bbc9cd'
  inverse-surface: '#e4e1ea'
  inverse-on-surface: '#303037'
  outline: '#859397'
  outline-variant: '#3c494c'
  surface-tint: '#2fd9f4'
  primary: '#8aebff'
  on-primary: '#00363e'
  primary-container: '#22d3ee'
  on-primary-container: '#005763'
  inverse-primary: '#006877'
  secondary: '#cebdff'
  on-secondary: '#381385'
  secondary-container: '#4f319c'
  on-secondary-container: '#bea8ff'
  tertiary: '#ffd0e3'
  on-tertiary: '#620040'
  tertiary-container: '#ffa6cf'
  on-tertiary-container: '#8f1e62'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#a2eeff'
  primary-fixed-dim: '#2fd9f4'
  on-primary-fixed: '#001f25'
  on-primary-fixed-variant: '#004e5a'
  secondary-fixed: '#e8ddff'
  secondary-fixed-dim: '#cebdff'
  on-secondary-fixed: '#21005e'
  on-secondary-fixed-variant: '#4f319c'
  tertiary-fixed: '#ffd8e7'
  tertiary-fixed-dim: '#ffafd3'
  on-tertiary-fixed: '#3d0026'
  on-tertiary-fixed-variant: '#85145a'
  background: '#131319'
  on-background: '#e4e1ea'
  surface-variant: '#35343b'
typography:
  display-lg:
    fontFamily: Plus Jakarta Sans
    fontSize: 48px
    fontWeight: '800'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Plus Jakarta Sans
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.3'
  headline-md:
    fontFamily: Plus Jakarta Sans
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.4'
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
  label-md:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.05em
  code-sm:
    fontFamily: JetBrains Mono
    fontSize: 13px
    fontWeight: '400'
    lineHeight: '1.5'
  headline-lg-mobile:
    fontFamily: Plus Jakarta Sans
    fontSize: 28px
    fontWeight: '700'
    lineHeight: '1.3'
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  unit: 4px
  xs: 4px
  sm: 8px
  md: 12px
  lg: 16px
  xl: 24px
  xxl: 32px
  gutter: 16px
  margin: 24px
---

## Brand & Style
The design system establishes a high-performance "Command Center" aesthetic for Site Reliability Engineers (SREs) and AI Operators. It balances extreme data density with a calm, engineering-grade clarity.

The visual style is **Corporate Modern with Glassmorphism accents**. It utilizes deep obsidian surfaces layered with translucent panels to create a sense of depth and focus. Subtle "aurora" radial glows in the background provide a premium, futuristic feel without distracting from critical metrics. The interface should feel like a high-precision instrument: responsive, reliable, and sophisticated.

## Colors
The color palette is optimized for long-duration monitoring. The **Aurora (Dark)** theme is the primary experience, reducing eye strain in low-light environments.

### Accent Strategy
- **Primary Cyan:** Used for active states, primary actions, and "Healthy" status indicators.
- **Gradient:** Reserved for high-level brand moments, AI-driven insights, or "Neural" processing states.
- **Severity Colors:** These follow a strict heatmap logic. In the dark theme, these colors are applied to text and icons, while their containers use a 14% opacity fill of the same hue to ensure legibility without overwhelming the dashboard.

### Background Orbs
In the Aurora theme, place large, soft radial gradients (15-20% opacity) in the background layer. Position them at the top-left (Cyan) and bottom-right (Violet) to break the monotony of the dark base.

## Typography
The typographic hierarchy prioritizes scanability. 
- **Plus Jakarta Sans** provides a modern, geometric feel for headers that differentiates the brand from standard utility tools. 
- **Inter** is used for all UI controls and body text to ensure maximum readability at small sizes.
- **JetBrains Mono** is mandatory for all technical data, including Trace IDs, Log timestamps, and JSON payloads, to ensure character alignment and a "developer-first" experience.

Use **uppercase labels** with slightly increased letter spacing for table headers and section overviews to create clear visual anchors.

## Layout & Spacing
The layout follows a **12-column fluid grid** for the main dashboard content. 

- **Sidebar:** Fixed at 240px (collapsed to 64px).
- **Density:** Use the 8px base unit for general layout, but switch to the 4px unit for dense data tables and log streams.
- **Margins:** 24px container padding on desktop, reducing to 16px on mobile.
- **Gutters:** 16px consistent spacing between dashboard widgets (cards).

Dashboards should use a "Masonry" or "Grid-stack" approach where widgets can span 3, 4, 6, or 12 columns depending on the complexity of the visualization.

## Elevation & Depth
Depth is communicated through **Tonal Layering** and **Glassmorphism**, rather than traditional heavy shadows.

1.  **Base Layer:** `#09090F` (The "Dark Room").
2.  **Surface Layer (Cards):** `#111118` with a 1px border of `rgba(255,255,255,0.06)`.
3.  **Elevated Layer (Modals/Popovers):** `#18181F` with a soft 20px blur shadow and a slightly brighter border (`0.1` opacity).

**Glassmorphism:** Navigation sidebars and header bars should use a backdrop filter (`blur(12px)`) and 80% opacity on their background color to allow the underlying background "orbs" to bleed through subtly.

## Shapes
The shape language is sophisticated and modern, moving away from harsh edges without becoming "bubbly."

- **Controls (Buttons, Inputs, Selects):** 8px radius for a precise, technical feel.
- **Standard Cards/Widgets:** 12px radius.
- **Main Content Panels/Layout Containers:** 16px radius for the outer corners.

Pill shapes are used exclusively for **Status Badges** and **Tags** to differentiate them from interactive buttons.

## Components
### Buttons
- **Primary:** Background `primary_cyan`, text dark. On hover, apply a cyan `box-shadow: 0 0 15px rgba(34, 211, 238, 0.4)`.
- **Secondary:** Ghost style with `1px` border of `primary_cyan`.
- **Tertiary:** Transparent background with `text_secondary`.

### Cards & Widgets
Every widget must feature a header with `label-md` text and an optional icon. The card background is `bg_surface` with a subtle 1px border. Use a slight inner glow (`inset 0 1px 0 rgba(255,255,255,0.05)`) to define the top edge.

### Inputs & Search
Fields use `bg_elevated` with a 1px border. Focus state changes border to `primary_cyan` with a subtle outer glow. The search bar in the log explorer should use JetBrains Mono for the query text.

### Badges (Severity)
Badges use a 14% opacity background of the semantic color with 100% opacity text. For example, a **Critical** badge is `#FB7185` text on `rgba(251, 113, 133, 0.14)` background.

### Charts
Series colors follow the defined palette: Cyan, Violet, Yellow, and Rose. Grid lines in charts should be extremely subtle (`rgba(255,255,255,0.03)`). Use Inter for axis labels and JetBrains Mono for precise value tooltips.