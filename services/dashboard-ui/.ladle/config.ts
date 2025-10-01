export default {
  previewDecorator: (Story) => {
    if (typeof window !== "undefined" && window.matchMedia) {
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
      const applyTheme = () => {
        document.documentElement.dataset.theme = mediaQuery.matches
          ? "dark"
          : "light";
      };
      applyTheme();
      mediaQuery.addEventListener("change", applyTheme);
    }
    return Story();
  },
};
