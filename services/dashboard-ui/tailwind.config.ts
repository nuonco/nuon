import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  safelist: [
    {
      pattern: /(bg|text)-(red|green|yellow)-(500|600)/,
    },
    {
      pattern: /(bg|text)-(slate)-(50|950)/,
    },
    {
      pattern: /bg-opacity-(25|50|75|85|95)/,
    },
    {
      pattern: /z-(10|20|30|40|50)/,
    },
    {
      pattern: /(bottom|top)-(0)/,
    },
    {
      pattern: /(m|p)-(auto)/,
    },
    {
      pattern: /(max|min)-(w|h)-(xs|sm|md|lg|xl|2xl)/,
    },
  ],
  theme: {
    extend: {
      gridAutoRows: {
        auto: '1fr',
      },
      gridTemplateColumns: {
        auto: 'repeat(auto-fill, minmax(18rem, 1fr))',
      },
    },
  },
  plugins: [],
}
export default config
