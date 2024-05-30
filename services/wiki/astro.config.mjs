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
      sidebar: [
        {
          label: "Nuon Wiki",
          //collapsed: true,
          items: [
            {
              label: "Home",
              link: "/",
            },
            {
              label: "Links",
              link: "links",
            },
          ],
        },
        {
          label: "Company",
          collapsed: true,
          autogenerate: {
            directory: "company",
          },
        },
        {
          label: "Sales",
          collapsed: true,
          autogenerate: {
            directory: "sales",
          },
        },
        {
          label: "Marketing",
          collapsed: true,
          autogenerate: {
            directory: "marketing",
          },
        },
        {
          label: "Product",
          collapsed: true,
          autogenerate: {
            directory: "product",
          },
        },
        {
          label: "Support",
          collapsed: true,
          autogenerate: {
            directory: "support",
          },
        },
      ],
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
