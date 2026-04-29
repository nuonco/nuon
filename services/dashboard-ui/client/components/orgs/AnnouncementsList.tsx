import { AnnouncementCard, type IAnnouncement } from './AnnouncementCard'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export interface IAnnouncementsList {
  announcements: IAnnouncement[]
  disableDismissMemory?: boolean
}

export const AnnouncementsList = ({ announcements, disableDismissMemory = false }: IAnnouncementsList) => {
  const visible = announcements.slice(0, 4)

  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Latest from our changelog
      </Text>
      {visible.map((announcement, i) => (
        <AnnouncementCard
          key={announcement.id}
          announcement={announcement}
          variant={i < 2 ? 'default' : 'compact'}
          disableDismissMemory={disableDismissMemory}
        />
      ))}
      <Text variant="body">
        <Link href="https://docs.nuon.co/updates/updates" isExternal>
          View changelog
          <Icon variant="ArrowSquareOutIcon" size={14} />
        </Link>
      </Text>
    </div>
  )
}
