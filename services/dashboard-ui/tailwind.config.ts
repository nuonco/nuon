import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      gridAutoRows:{
        "auto": "1fr"
      },
      gridTemplateColumns: {
        "auto": "repeat(auto-fill, minmax(18rem, 1fr))"
      }
    }
  },
  plugins: [],
};
export default config;
