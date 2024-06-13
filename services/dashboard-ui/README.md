This is a [Next.js](https://nextjs.org/) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

## Getting Started

First, you'll need to create an `.env.local` file. You can copy the `.example.env.local` file and update the required values.

``` bash
cp .example.env.local .env.local
```

Install the project dependencies using `npm install`.

Then, run the development server:

```bash
npm run dev
```

Open [http://localhost:4000](http://localhost:4000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/basic-features/font-optimization) to automatically optimize and load Inter, a custom Google Font.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.

## Integrity checks

We aim to make changes quickly and safely, using integrity checks that run in GitHub actions when you create a PR. Most text editors can be configured to format and lint code automatically and you can run these checks locally using either Earthly (e.g. `earthly +test --image_tag=local --repo=mono`) or `npm`.


- lint (default next eslint): `npm run lint`

- typecheck (typescript): `npm run tsc`, `npm run tsc -- --watch`

- fmt (prettier): `npm run fmt`, `npm run fmt -- --write` (to format code in src)

- unit test (vitest + react-testing-lib + msw): `npm run test`, `npm run test -- --watch`
