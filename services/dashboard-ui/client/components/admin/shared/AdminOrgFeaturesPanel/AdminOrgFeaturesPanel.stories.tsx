export default {
  title: 'Admin/AdminOrgFeaturesPanel',
}

import { AdminOrgFeaturesPanel } from './AdminOrgFeaturesPanel'

export const Default = () => (
  <AdminOrgFeaturesPanel
    org={{ features: { 'feature-a': true, 'feature-b': false } } as any}
    orgId="org-1"
    featuresList={[
      { name: 'feature-a', description: 'An org-toggleable feature.' },
      { name: 'feature-b', description: 'Another org-toggleable feature.' },
      {
        name: 'feature-c',
        description: 'Pinned on by the deployment config.',
        forced: true,
      },
    ]}
    isLoading={false}
    isSubmitting={false}
    onSubmit={(e) => e.preventDefault()}
  />
)

export const ScrollingList = () => (
  <AdminOrgFeaturesPanel
    isVisible
    org={{ features: { 'feature-3': true, 'feature-9': true } } as any}
    orgId="org-1"
    featuresList={Array.from({ length: 30 }).map((_, index) => ({
      name: `feature-${index}`,
      description:
        'A longer description so the list overflows the panel and scrolls beneath the footer.',
    }))}
    isLoading={false}
    isSubmitting={false}
    onSubmit={(e) => e.preventDefault()}
  />
)

export const Loading = () => (
  <AdminOrgFeaturesPanel
    org={{} as any}
    orgId="org-1"
    featuresList={[]}
    isLoading={true}
    isSubmitting={false}
    onSubmit={(e) => e.preventDefault()}
  />
)
