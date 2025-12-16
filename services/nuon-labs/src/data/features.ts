export type FeatureStatus = 'alpha' | 'beta' | 'stable';

export interface LabFeature {
  id: string;
  title: string;
  description: string;
  backgroundImage?: string;
  href: string;
  isExternal: boolean;
  status: FeatureStatus;
}

export const labFeatures: LabFeature[] = [
  {
    id: 'customer-dashboard',
    title: 'Customer Dashboard',
    description: 'Purpose-built dashboard for customer to install your apps and approve updates.',
    href: 'https://vendor.inl0qjpbg8hn5e25ebmcjzmwh2.nuon.run/',
    isExternal: true,
    status: 'alpha',
  },
  {
    id: 'nuon-tui',
    title: 'TUI for Nuon CLI',
    description: 'Contextual and full-window TUIs for common Nuon workflows.',
    href: 'https://docs.nuon.co/cli#nuon-preview-features',
    isExternal: true,
    status: 'alpha',
  },
];
