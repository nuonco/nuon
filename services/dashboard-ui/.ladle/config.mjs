export default {
  stories: [
    "client/**/*.stories.{tsx,jsx,ts,js}",
    "!client/lite/**",
  ],
  port: 61000,
  previewPort: 61001,
  storyOrder: (stories) => {
    const rank = (id) => {
      const root = id.split("--")[0]
      if (root === "ui") return 0
      if (root === "features") return 1
      if (root === "views") return 2
      if (root === "playground") return 3
      return 4
    }
    return [...stories].sort((a, b) => rank(a) - rank(b) || a.localeCompare(b))
  },
};
