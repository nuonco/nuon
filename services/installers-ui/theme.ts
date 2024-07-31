const colors = require("tailwindcss/colors");
import { generateColorRange } from "./common/generateColors";

const primary = generateColorRange("#4E17FF");

export default {
  darkMode: ["selector"],
  forceDarkMode: true,
  theme: {
    extend: {
      colors: {
        black: colors.black,
        white: colors.white,
        primary: primary,
        "card-text-color": colors.white,
        "card-border-color": colors.black,
        "card-background": "rgb(255 255 255 / 5%)",
        "card-text-color-dark": colors.black,
        "card-border-color-dark": colors.white,
        "card-background-dark": "rgb(0,0,0 / 5%)",
        "accordion-header-color": colors.black,
        "accordion-header-background": primary["500"],
        "accordion-header-active-color": colors.black,
        "accordion-header-active-color-dark": colors.gray["100"],
        "accordion-header-active-background": colors.blue["100"],
        "accordion-header-active-background-dark": colors.blue["900"],
        "step-active-border-color": colors.orange["400"],
        "step-complete-border-color": colors.green["500"],
        "button-text-color": colors.gray["50"],
        "button-border-color": colors.indigo["900"],
        "button-bg-color": colors.green["500"],
        "button-bg-hover": primary["600"],
        "button-bg-focus": primary["600"],
        "button-bg-active": primary["700"],
      },
      boxShadow: {
        "card-shadow": "4px 4px 0 0 black, 0 1px 2px -1px black",
        "card-shadow-dark": "4px 4px 0 0 white, 0 1px 2px -1px white",
        "button-shadow": "3px 3px 0 0 black, 0 1px 2px -1px black",
        "button-shadow-dark": "3px 3px 0 0 white, 0 1px 2px -1px white",
      },
      borderWidth: {
        "card-border-width": "1px",
        "button-border-width": "0px",
      },
      borderRadius: {
        "card-radius": "0.25rem",
        "button-radius": "0.25rem",
        "video-container-radius": "0.5rem",
      },
    },
  },
};
