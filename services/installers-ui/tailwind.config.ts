import type { Config } from "tailwindcss";
import withMT from "@material-tailwind/react/utils/withMT";

import theme from "./theme";

const config: Config = {
  content: [
    "./common/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./node_modules/react-tailwindcss-select/dist/index.esm.js",
  ],
  safelist: [
    {
      pattern: /(bg|text)-(red|green|yellow)-(500|600)/,
    },
  ],
  darkMode: theme.darkMode as Config["darkMode"],
  theme: theme.theme,
  plugins: [require("@tailwindcss/typography")],
};
export default withMT(config);
