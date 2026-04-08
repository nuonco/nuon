export default {
  title: 'Orgs/AnnouncementCard',
}

import { AnnouncementCard } from './AnnouncementCard'
import type { IAnnouncement } from './AnnouncementCard'

const mockAnnouncement: IAnnouncement = {
  id: 'ann-1',
  title: 'Nuon 2.0 is here',
  date: 'January 15, 2025',
  description: 'We are excited to announce the release of Nuon 2.0, packed with new features and improvements.',
  ctaText: 'Read more',
  ctaUrl: 'https://nuon.co/blog',
  dismissible: true,
}

export const Default = () => <AnnouncementCard announcement={mockAnnouncement} />

export const WithImage = () => (
  <AnnouncementCard
    announcement={{
      ...mockAnnouncement,
      image: 'https://via.placeholder.com/800x400',
    }}
  />
)

export const NonDismissible = () => (
  <AnnouncementCard
    announcement={{ ...mockAnnouncement, dismissible: false }}
  />
)
