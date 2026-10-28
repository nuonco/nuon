import type { TInstallShare } from './types'
import type { TSharedInstallRow } from './SharedInstallsTable/SharedInstallsTable'

export const mockShares: TInstallShare[] = [
  {
    id: 'share-1',
    email: 'ops@acme.example.com',
    scopes: ['readme', 'resources', 'approvals'],
    status: 'accepted',
    shared_at: '2026-10-20T14:22:00Z',
    install_id: 'inst-1',
    install_name: 'Acme production',
    app_name: 'Payments',
  },
  {
    id: 'share-2',
    email: 'platform@acme.example.com',
    scopes: ['readme', 'resources', 'operations'],
    status: 'pending',
    shared_at: '2026-10-26T09:10:00Z',
    install_id: 'inst-1',
    install_name: 'Acme production',
    app_name: 'Payments',
  },
  {
    id: 'share-3',
    email: 'viewer@globex.example.com',
    scopes: ['readme'],
    status: 'accepted',
    shared_at: '2026-10-12T18:05:00Z',
    install_id: 'inst-2',
    install_name: 'Globex staging',
    app_name: 'Analytics',
  },
]

export const mockSharedInstallRows: TSharedInstallRow[] = [
  {
    installId: 'inst-1',
    installName: 'Acme production',
    installHref: '/org-1/installs/inst-1',
    appName: 'Payments',
    shares: mockShares.filter((share) => share.install_id === 'inst-1'),
    sharedAt: '2026-10-26T09:10:00Z',
  },
  {
    installId: 'inst-2',
    installName: 'Globex staging',
    installHref: '/org-1/installs/inst-2',
    appName: 'Analytics',
    shares: mockShares.filter((share) => share.install_id === 'inst-2'),
    sharedAt: '2026-10-12T18:05:00Z',
  },
]
