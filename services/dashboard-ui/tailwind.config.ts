import type { Config } from 'tailwindcss'
import colors from 'tailwindcss/colors'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
    './node_modules/react-tailwindcss-select/dist/index.esm.js',
  ],
  safelist: [
    {
      pattern: /(bg|text)-(red|green|yellow)-(500|600)/,
    },
    {
      pattern: /(bg|text)-(slate)-(50|950)/,
    },
    {
      pattern: /(bg|text)-(active|light|dark)/,
    },
    {
      pattern: /bg-opacity-(25|50|75|85|95)/,
    },
    {
      pattern: /z-(10|20|30|40|50)/,
    },
    {
      pattern: /(bottom|top|left|right)-(0|2|4|6|8|10)/,
    },
    {
      pattern: /(m|p)-(auto)/,
    },
    {
      pattern: /(max|min)-(w|h)-(xs|sm|md|lg|xl|2xl)/,
    },
  ],
  theme: {
    colors: {
      ...colors,
      dark: '#100E16',
      light: '#ffffff',
      active: '#8040BF',
      primary: {
        50: '#FCFAFF',
        100: '#F6F0FF',
        200: '#F2E5FF',
        300: '#E5D0FB',
        400: '#C494F4',
        500: '#AD71EA',
        // default
        600: '#8040BF',
        700: '#7339AC',
        800: '#662F9D',
        900: '#4C2277',
        950: '#2E0E4E',
      },
      blue: {
        50: '#FAFBFF',
        100: '#EDF2FF',
        200: '#E5EEFF',
        300: '#CDDDFF',
        400: '#8DB0FB',
        500: '#6792F4',
        // default
        600: '#3062D4',
        700: '#2759CD',
        800: '#1E50C0',
        900: '#113997',
        950: '#05205E',
      },
      green: {
        50: '#F4FBF7',
        100: '#E6F9EF',
        200: '#D8F8E7',
        300: '#C6F1DA',
        400: '#75CC9E',
        500: '#4AA578',
        // default
        600: '#1D7C4D',
        700: '#1E714A',
        800: '#196742',
        900: '#0E4E30',
        950: '#062D1B',
      },
      orange: {
        50: '#FFF5EB',
        100: '#FFF0E0',
        200: '#FFE8D1',
        300: '#FFD4A8',
        400: '#FEB872',
        500: '#F6A351',
        // default
        600: '#F59638',
        700: '#B4610E',
        800: '#A05C1C',
        900: '#7A4510',
        950: '#482909',
      },
      red: {
        50: '#FEF2F2',
        100: '#FEE2E2',
        200: '#FECACA',
        300: '#FCA5A5',
        400: '#F87171',
        500: '#EF4444',
        // default
        600: '#DC2626',
        700: '#B91C1C',
        800: '#991B1B',
        900: '#7F1D1D',
        950: '#450A0A',
      },
      'cool-grey': {
        50: '#F9F9FB',
        100: '#F0F3F5',
        200: '#EAEDF0',
        300: '#DEE3E7',
        // default
        400: '#CFD6DD',
        500: '#9EA8B3',
        600: '#555F6D',
        700: '#4A545E',
        800: '#3A424A',
        900: '#272E35',
        950: '#1B242C',
      },
      'dark-grey': {
        50: '#121212',
        // default
        100: '#141217',
        200: '#19171C',
        300: '#1D1B20',
        400: '#222025',
        500: '#27252A',
        600: '#2C2A2E',
        700: '#302E33',
        800: '#353337',
        900: '#3A383C',
        950: '#3E3D41',
      },
    },
    fontFamily: {
      sans: ['var(--font-geist-sans)'],
      mono: ['var(--font-geist-mono)'],
    },
    fontSize: {
      xs: '8px',
      sm: '12px',
      base: '14px',
      lg: '16px',
      xl: '18px',
    },
    fontWeight: {
      normal: '400',
      medium: '500',
      semibold: '600',
    },
    letterSpacing: {
      normal: '0',
      wide: '0.2px',
    },
    lineHeight: {
      none: '1',
      tight: '12px',
      normal: '16px',
      relaxed: '20px',
      loose: '24px',
    },
    extend: {
      gridAutoRows: {
        auto: '1fr',
      },
      gridTemplateColumns: {
        auto: 'repeat(auto-fill, minmax(18rem, 1fr))',
      },
      width: {
        inherit: 'inherit',
      },
    },
  },
  plugins: [require('@tailwindcss/typography')],
}
export default config
