This is a [Next.js](https://nextjs.org/) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app). 

## Getting Started

### Prerequisites

The dashboard requires Auth0 for authentication. You can create a free Auth0 account at [auth0.com](https://auth0.com) and set up a new application to get the required credentials. See the `.env.local.example` file for the required Auth0 environment variables.

**Note for External Contributors**: For most development use cases, you may want to work with the API or CLI instead if you don't need the dashboard UI. Alternatively, you can use the hosted dashboard at [https://app.nuon.co](https://app.nuon.co).

### Setup

First, create an `.env.local` file with your Auth0 configuration:

```bash
cp .env.local.example .env.local
# Edit .env.local to add Auth0 environment variables
```

Install the project dependencies:

```bash
npm install
```

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
