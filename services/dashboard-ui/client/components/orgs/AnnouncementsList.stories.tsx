export default {
  title: 'Orgs/AnnouncementsList',
}

import { AnnouncementsList } from './AnnouncementsList'
import type { IAnnouncement } from './AnnouncementCard'

const mockAnnouncements: IAnnouncement[] = [
  {
    id: 'story-1',
    title: 'Multi-cloud support',
    date: 'Apr 28, 2026',
    description: 'Deploy to Azure AKS and GCP GKE clusters alongside AWS EKS.',
    image: 'https://via.placeholder.com/800x400',
    imageDark: 'https://via.placeholder.com/800x400',
    ctaText: 'Learn more',
    ctaUrl: 'https://docs.nuon.co/updates/038-multi-cloud',
    dismissible: false,
  },
  {
    id: 'story-2',
    title: 'Workflow approvals v2',
    date: 'Apr 21, 2026',
    description: 'Redesigned approval flows with bulk approve and Slack notifications.',
    image: 'https://via.placeholder.com/800x400',
    ctaText: 'Learn more',
    ctaUrl: 'https://docs.nuon.co/updates/037-approvals-v2',
    dismissible: true,
  },
  {
    id: 'story-3',
    title: 'Runner health dashboard',
    date: 'Apr 14, 2026',
    description: 'Monitor runner uptime and resource usage from a single view.',
    ctaText: 'Learn more',
    ctaUrl: 'https://docs.nuon.co/updates/036-runner-health',
    dismissible: true,
  },
  {
    id: 'story-4',
    title: 'Install drift detection',
    date: 'Apr 7, 2026',
    description: 'Automatic drift detection alerts when install state diverges.',
    ctaText: 'Try now',
    ctaUrl: 'https://docs.nuon.co/updates/035-drift-detection',
    dismissible: true,
  },
  {
    id: 'story-5',
    title: 'CLI extensions',
    date: 'Mar 31, 2026',
    description: 'This one should not appear since we cap at 4.',
    ctaText: 'Learn more',
    ctaUrl: 'https://docs.nuon.co/updates/034-cli-extensions',
    dismissible: true,
  },
]

export const Default = () => (
  <div style={{ maxWidth: 400 }}>
    <AnnouncementsList announcements={mockAnnouncements} disableDismissMemory />
  </div>
)

export const SingleAnnouncement = () => (
  <div style={{ maxWidth: 400 }}>
    <AnnouncementsList announcements={[mockAnnouncements[0]]} disableDismissMemory />
  </div>
)

export const TwoAnnouncements = () => (
  <div style={{ maxWidth: 400 }}>
    <AnnouncementsList announcements={mockAnnouncements.slice(0, 2)} disableDismissMemory />
  </div>
)
