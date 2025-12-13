import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    fontFamily: {
      sans: ['var(--font-inter)', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      mono: ['var(--font-hack)', 'ui-monospace', 'monospace'],
    },
  },
  plugins: [],
}

export default config
