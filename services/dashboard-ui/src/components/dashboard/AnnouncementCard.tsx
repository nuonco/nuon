'use client'

import { useState } from 'react'
import { ArrowRight } from '@phosphor-icons/react'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'

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
}: {
  announcement: IAnnouncement
}) => {
  const [isDismissed, setIsDismissed] = useState(false)

  if (isDismissed) return null

  return (
    <div className="border rounded-lg overflow-hidden bg-white dark:bg-dark-grey-800">
      {announcement.image && (
        <div className="relative w-full aspect-video bg-cool-grey-100 dark:bg-dark-grey-700 overflow-hidden">
          {/* Light mode image */}
          <img
            src={announcement.image}
            alt={announcement.title}
            className="w-full h-full object-cover dark:hidden"
          />
          {/* Dark mode image */}
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
          <Button
            variant="secondary"
            size="sm"
            href={announcement.ctaUrl}
          >
            {announcement.ctaText}
            <ArrowRight size={14} weight="bold" />
          </Button>
          {announcement.dismissible && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsDismissed(true)}
            >
              Dismiss
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
