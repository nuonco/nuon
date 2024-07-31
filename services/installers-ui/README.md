The Nuon installer UI is a [Next.js](https://nextjs.org/) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

## Getting Started

> [!NOTE]
> We recommend forking this repository to stay current with the latest version.

First, create a `.env.local` file then add your Nuon access token and org ID.

```bash
NUON_API_TOKEN='...'
NUON_ORG_ID='...'
NUON_INSTALLER_ID='...'
```

You can access this information using the [Nuon CLI](https://docs.nuon.co/cli) if you're unsure.

```bash
nuon orgs print-config
```

Then, create a theme config file by copying the example one we have provided. You can update this file to customize the theme.

```bash
cp theme.ts.example theme.ts
```

Finally, install the dependencies and run the development server:

```bash
npm i && npm run dev
# or
yarn && yarn dev
# or
pnpm i && pnpm dev
# or
bun i && bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

Under the hood the Nuon installer UI is using the [Nuon REST API](https://docs.nuon.co/api-ref/getting-started) and can be customized to your experience needs.

This project uses [`next/font`](https://nextjs.org/docs/basic-features/font-optimization) to automatically optimize and load Inter, a custom Google Font.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js/) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/deployment) for more details.

## Self-hosted with Docker image

Next.js can be deployed to any hosting provider that supports Docker containers.

With Docker installed on your machine you can build and run your Installer UI in a Docker container. First you'll need to copy your env values to a `.env` file then build and run the Docker container.

```bash
# copy .env.local
cp .env.local .env
# build container
docker build -t installer-docker .
# run container
docker run -p 3000:3000 installer-docker
```
