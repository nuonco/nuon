# Nuon Labs Service

The **Nuon Labs** service is an experimental features showcase website hosted at `https://labs.nuon.co`. It displays cutting-edge features and tools being developed by the Nuon team.

## Service Overview

This is a Next.js application that showcases experimental Nuon features with a 3D particle-based logo background rendered using Three.js. The site uses the Stratus design system from the dashboard-ui for consistent branding.

## Architecture

- **Framework**: Next.js 15+ with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4.1 with Stratus design tokens
- **3D Graphics**: Three.js with @react-three/fiber and @react-three/drei
- **Icons**: Phosphor Icons (@phosphor-icons/react)

## Project Structure

### Core Files
- `next.config.mjs` - Next.js configuration
- `tailwind.config.ts` - Tailwind CSS with Stratus tokens
- `package.json` - Dependencies and scripts
- `service.yml` - nuonctl local development configuration
- `Dockerfile` - Container build configuration

### Key Directories

#### `/src/app/` - Next.js App Router
- `layout.tsx` - Root layout with Inter and JetBrains Mono fonts
- `page.tsx` - Landing page with feature grid and 3D background
- `globals.css` - Stratus design system CSS variables

#### `/src/components/` - React Components
- `common/` - Stratus design system components (Button, Card, Text)
- `ThreeBackground/` - Three.js particle logo background
- `Header.tsx` - Site header with Nuon logo and navigation
- `Footer.tsx` - Site footer with links
- `FeatureCard.tsx` - Individual feature display card
- `FeatureGrid.tsx` - 2-column grid of feature cards

#### `/src/data/` - Static Content
- `features.ts` - Experimental feature definitions

#### `/src/utils/` - Utility Functions
- `classnames.ts` - cn() utility for Tailwind class composition

## Key Features

### 3D Particle Background
- Three.js particle system rendering the Nuon logo
- Uses logo gradient colors: pink (#F72585) -> purple (#3A00FF) -> cyan (#4CC9F0)
- Breathing and floating animation effects
- Client-side only rendering with dynamic import

### Feature Cards
- 2-column responsive grid
- Status badges (alpha, beta, stable)
- Hover animations with gradient effects
- External links to try features

### Stratus Design System
- Purple primary color palette
- Inter font (sans) and JetBrains Mono (mono)
- Dark theme by default
- Consistent with dashboard-ui styling

## Development

### Setup
```bash
cd services/nuon-labs
npm install
npm run dev
```

### Using nuonctl
```bash
./run-nuonctl.sh services dev --dev nuon-labs
```

### Key Scripts
- `npm run dev` - Development server on port 4100
- `npm run build` - Production build
- `npm run start` - Start production server

## Adding New Features

Edit `/src/data/features.ts` to add new experimental features:

```typescript
{
  id: 'feature-id',
  title: 'Feature Title',
  description: 'Feature description text.',
  href: 'https://app.nuon.co/path',
  isExternal: true,
  status: 'alpha' | 'beta' | 'stable',
}
```

## Port Configuration
- Local development: 4100
- Health check: http://localhost:4100

## Deployment

The site is containerized with Docker and can be deployed to:
- Kubernetes via Helm charts
- Vercel for static hosting
- Any container orchestration platform

DNS configuration should point `labs.nuon.co` to the deployed service.
