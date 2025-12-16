/**
 * Nuon Labs Design System
 * 
 * A dark, terminal-inspired aesthetic with monospace typography
 * and subtle orange accents.
 */

// =============================================================================
// COLORS
// =============================================================================

export const colors = {
  // Backgrounds
  background: {
    primary: '#000000',      // Main page background
    elevated: '#111111',     // Cards, terminal body
    surface: '#191919',      // Terminal header, subtle elevation
    input: '#0d0d0d',        // Input areas
  },

  // Borders
  border: {
    default: '#2a2a2a',      // Standard borders
    subtle: '#1a1a1a',       // Subtle dividers
  },

  // Text
  text: {
    primary: '#e5e5e5',      // Main text, headings
    secondary: '#a3a3a3',    // Secondary text, command output
    tertiary: '#808080',     // Muted text, dates
    muted: '#666666',        // Labels, placeholders
    disabled: '#525252',     // Disabled states, hints
    faint: '#404040',        // Very subtle text (copyright)
  },

  // Brand / Accent
  accent: {
    primary: '#f97316',      // Orange - prompts, highlights, bullets
    blue: '#006CFF',         // Blue - background gradient
  },

  // Semantic
  semantic: {
    success: '#28c840',      // Green - success states
    warning: '#febc2e',      // Yellow - warnings
    error: '#ff5f57',        // Red - errors
  },

  // Terminal dots
  terminal: {
    close: '#ff5f57',
    minimize: '#febc2e',
    maximize: '#28c840',
  },
} as const

// =============================================================================
// TYPOGRAPHY
// =============================================================================

export const fonts = {
  // Primary - used site-wide for body text
  mono: 'var(--font-space-mono)',
  
  // Accent - used for headings and titles
  heading: 'var(--font-ibm-plex-mono)',
  
  // Additional available fonts
  hack: 'var(--font-hack)',           // JetBrains Mono
  serif: 'var(--font-serif)',         // Instrument Serif
  sans: 'var(--font-inter)',          // Inter
} as const

export const fontSizes = {
  xs: '11px',      // Table headers, hints
  sm: '12px',      // Labels, small text
  base: '14px',    // Body text
  lg: '16px',      // Slightly larger body
  xl: '18px',      // Subheadings
  '2xl': '24px',   // Section titles
  '3xl': '32px',   // Page titles
  '4xl': '40px',   // Hero (md)
  '5xl': '48px',   // Hero (lg)
  '6xl': '64px',   // Hero (desktop)
  '7xl': '80px',   // Hero (large desktop)
  '8xl': '96px',   // Hero (xl desktop)
} as const

// =============================================================================
// SPACING
// =============================================================================

export const spacing = {
  container: '80%',          // Main content width
  px: {
    sm: '1rem',              // 16px - mobile padding
    md: '1.5rem',            // 24px - default padding
    lg: '2rem',              // 32px - larger sections
  },
  py: {
    sm: '0.75rem',           // 12px
    md: '1.25rem',           // 20px
    lg: '2rem',              // 32px
    xl: '4rem',              // 64px - large sections
  },
  gap: {
    xs: '0.25rem',           // 4px
    sm: '0.5rem',            // 8px
    md: '1rem',              // 16px
    lg: '1.5rem',            // 24px
    xl: '2rem',              // 32px
  },
} as const

// =============================================================================
// BORDERS & RADIUS
// =============================================================================

export const borders = {
  default: `1px solid ${colors.border.default}`,
  subtle: `1px solid ${colors.border.subtle}`,
  radius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    full: '9999px',
  },
} as const

// =============================================================================
// ANIMATIONS
// =============================================================================

export const animations = {
  duration: {
    fast: '200ms',
    normal: '300ms',
    slow: '500ms',
    slower: '700ms',
    slowest: '1000ms',
  },
  easing: {
    default: 'ease',
    out: 'ease-out',
    in: 'ease-in',
    inOut: 'ease-in-out',
    // Custom cubic-bezier for intro animations
    outCubic: 'cubic-bezier(0.33, 1, 0.68, 1)',
  },
} as const

// =============================================================================
// COMPONENT STYLES
// =============================================================================

export const components = {
  // Terminal window
  terminal: {
    background: colors.background.elevated,
    border: borders.default,
    borderRadius: borders.radius.xl,
    headerBg: colors.background.surface,
    inputBg: colors.background.input,
    height: '550px',
  },

  // Buttons / Interactive elements
  button: {
    primary: {
      bg: colors.accent.primary,
      text: '#ffffff',
      hoverBg: '#ea580c',
    },
    ghost: {
      text: colors.text.muted,
      hoverText: colors.text.primary,
    },
  },

  // Table / List items
  table: {
    headerText: colors.text.muted,
    headerSize: fontSizes.xs,
    rowBorder: borders.default,
    hoverBg: colors.background.elevated,
  },

  // Status badges
  badge: {
    border: borders.default,
    text: colors.text.tertiary,
    fontSize: fontSizes.xs,
  },
} as const

// =============================================================================
// TAILWIND CLASS HELPERS
// =============================================================================

/**
 * Common Tailwind class combinations for quick styling
 */
export const tw = {
  // Text styles
  textPrimary: 'text-[#e5e5e5]',
  textSecondary: 'text-[#a3a3a3]',
  textMuted: 'text-[#666666]',
  textAccent: 'text-[#f97316]',

  // Background styles
  bgElevated: 'bg-[#111111]',
  bgSurface: 'bg-[#191919]',
  bgInput: 'bg-[#0d0d0d]',

  // Border styles
  borderDefault: 'border border-[#2a2a2a]',
  borderSubtle: 'border border-[#1a1a1a]',

  // Common patterns
  container: 'w-[80%] mx-auto',
  card: 'bg-[#111111] border border-[#2a2a2a] rounded-xl',
  
  // Transitions
  transitionFast: 'transition-all duration-200',
  transitionNormal: 'transition-all duration-300',
  transitionSlow: 'transition-all duration-1000',

  // Fade in animation
  fadeIn: 'opacity-100 translate-y-0',
  fadeOut: 'opacity-0 translate-y-8',
} as const

// =============================================================================
// USAGE EXAMPLES
// =============================================================================

/**
 * Example usage in components:
 * 
 * import { colors, fonts, tw } from '@/lib/design-system'
 * 
 * // Using color tokens
 * <div style={{ color: colors.text.primary }}>Hello</div>
 * 
 * // Using font tokens  
 * <h1 style={{ fontFamily: fonts.heading }}>Title</h1>
 * 
 * // Using Tailwind helpers
 * <div className={`${tw.card} ${tw.transitionNormal}`}>Card</div>
 * 
 * // Combining patterns
 * <div className={`${tw.container} ${tw.textPrimary}`}>Content</div>
 */

