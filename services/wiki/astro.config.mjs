import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwind from "@astrojs/tailwind";

import node from "@astrojs/node";

// https://astro.build/config
export default defineConfig({
  vite: {
    resolve: {
      preserveSymlinks: true,
    },
  },
  integrations: [
    starlight({
      title: "Nuon Wiki",
      social: {
        github: "https://github.com/nuonco",
        twitter: "https://twitter.com/nuoninc",
        youtube: "https://www.youtube.com/channel/UC5zHWPIfIIfgpPMNo_gonPw",
        linkedin: "https://www.linkedin.com/company/nuonco/",
      },
      customCss: ["./src/tailwind.css"],
      pagination: false,
      // sidebar: [
      //   {
      //     label: "Company",
      //     autogenerate: {
      //       directory: "company",
      //     },
      //   },
      //   {
      //     label: "Sales",
      //     autogenerate: {
      //       directory: "sales",
      //     },
      //   },
      //   {
      //     label: "Marketing",
      //     autogenerate: {
      //       directory: "marketing",
      //     },
      //   },
      //   {
      //     label: "Product",
      //     autogenerate: {
      //       directory: "product",
      //     },
      //   },
      //   {
      //     label: "Legacy Wiki",
      //     autogenerate: {
      //       directory: "../../wiki",
      //     },
      //   },
      // ],
    }),
    tailwind({
      applyBaseStyles: false,
    }),
  ],
  output: "server",
  adapter: node({
    mode: "standalone",
  }),
});
