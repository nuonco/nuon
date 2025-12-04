import type { Config } from 'tailwindcss'
// import colors from 'tailwindcss/colors'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
    './node_modules/react-tailwindcss-select/dist/index.esm.js',
  ],
  theme: {
    fontFamily: {
      sans: ['var(--font-inter)', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      mono: ['var(--font-hack)', 'ui-monospace', 'monospace'],
    },
  },
  // plugins: [require('@tailwindcss/typography')],
}
export default config
