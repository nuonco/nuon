import { useCallback, useRef, useState } from 'react'
import './AnnouncementCard.css'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

const DISMISSED_KEY = 'nuon:dismissed-announcements'

function getDismissedIds(): string[] {
  try {
    return JSON.parse(localStorage.getItem(DISMISSED_KEY) || '[]')
  } catch {
    return []
  }
}

function persistDismiss(id: string) {
  const ids = getDismissedIds()
  if (!ids.includes(id)) {
    localStorage.setItem(DISMISSED_KEY, JSON.stringify([...ids, id]))
  }
}

export interface IAnnouncement {
  id: string
  title: string
  date: string
  description: string
  image?: string
  imageDark?: string
  ctaText: string
  ctaUrl: string
  dismissible?: boolean
}

export const AnnouncementCard = ({
  announcement,
  variant = 'default',
  disableDismissMemory = false,
}: {
  announcement: IAnnouncement
  variant?: 'default' | 'compact'
  disableDismissMemory?: boolean
}) => {
  const [phase, setPhase] = useState<'visible' | 'sliding' | 'collapsing' | 'removed'>(
    !disableDismissMemory && getDismissedIds().includes(announcement.id) ? 'removed' : 'visible'
  )
  const wrapperRef = useRef<HTMLDivElement>(null)

  const handleDismiss = useCallback(() => {
    persistDismiss(announcement.id)
    setPhase('sliding')
    setTimeout(() => {
      if (wrapperRef.current) {
        wrapperRef.current.style.height = `${wrapperRef.current.offsetHeight}px`
        wrapperRef.current.offsetHeight // force reflow
        wrapperRef.current.style.height = '0px'
      }
      setPhase('collapsing')
      setTimeout(() => setPhase('removed'), 200)
    }, 150)
  }, [announcement.id])

  if (phase === 'removed') return null

  const slidingClass = phase === 'sliding' ? 'announcement-slide-out' : ''
  const hiddenClass = phase === 'collapsing' ? 'announcement-hidden' : ''
  const collapsingClass = phase === 'collapsing' ? 'announcement-collapse' : ''

  if (variant === 'compact') {
    return (
      <div ref={wrapperRef} className={`announcement-wrapper ${collapsingClass}`}>
        <div className={`border rounded-lg overflow-hidden bg-white dark:bg-dark-grey-800 p-4 flex flex-col gap-2 relative ${slidingClass} ${hiddenClass}`}>
          {announcement.dismissible && (
            <Button
              variant="ghost"
              size="sm"
              className="absolute top-2 right-2 !p-1"
              onClick={handleDismiss}
            >
              <Icon variant="XIcon" size={14} />
            </Button>
          )}
          <Text variant="subtext" theme="neutral">
            {announcement.date}
          </Text>
          <Text variant="base" weight="strong">
            <Link href={announcement.ctaUrl} isExternal className="!text-inherit hover:!text-primary-600 dark:hover:!text-primary-400">
              {announcement.title}
              <Icon variant="ArrowSquareOutIcon" size={14} />
            </Link>
          </Text>
        </div>
      </div>
    )
  }

  return (
    <div ref={wrapperRef} className={`announcement-wrapper ${collapsingClass}`}>
      <div className={`border rounded-lg overflow-hidden bg-white dark:bg-dark-grey-800 ${slidingClass} ${hiddenClass}`}>
        {announcement.image && (
          <div className="relative w-full aspect-video bg-cool-grey-100 dark:bg-dark-grey-700 overflow-hidden">
            <img
              src={announcement.image}
              alt={announcement.title}
              className="w-full h-full object-cover dark:hidden"
            />
            {announcement.imageDark && (
              <img
                src={announcement.imageDark}
                alt={announcement.title}
                className="w-full h-full object-cover hidden dark:block"
              />
            )}
          </div>
        )}
        <div className="p-5 flex flex-col gap-3">
          <Text variant="base" weight="strong">
            {announcement.title}
          </Text>
          <Text variant="subtext" theme="neutral">
            {announcement.date}
          </Text>
          <Text variant="body" theme="neutral">
            {announcement.description}
          </Text>
          <div className="flex items-center justify-between gap-3 mt-2">
            <Button variant="secondary" size="sm" href={announcement.ctaUrl}>
              {announcement.ctaText}
              <Icon variant="ArrowRightIcon" size={14} />
            </Button>
            {announcement.dismissible && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDismiss}
              >
                Dismiss
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
