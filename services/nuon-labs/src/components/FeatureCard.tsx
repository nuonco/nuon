'use client'

import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { cn } from '@/utils/classnames'
import { ArrowRight } from '@phosphor-icons/react'
import type { LabFeature, FeatureStatus } from '@/data/features'

const STATUS_STYLES: Record<FeatureStatus, string> = {
  alpha: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
  beta: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  stable: 'bg-green-500/20 text-green-400 border-green-500/30',
}

export const FeatureCard = ({ feature }: { feature: LabFeature }) => {
  const { title, description, href, status, backgroundImage } = feature

  return (
    <div
      className={cn(
        'group relative flex flex-col justify-between',
        'p-6 rounded-xl border border-white/10 bg-dark-grey-800/50',
        'backdrop-blur-sm transition-all duration-300',
        'hover:border-primary-500/50 hover:bg-dark-grey-700/50',
        'hover:shadow-lg hover:shadow-primary-500/10',
        'min-h-[240px]'
      )}
      style={
        backgroundImage
          ? {
              backgroundImage: `linear-gradient(to bottom, rgba(20, 18, 23, 0.9), rgba(20, 18, 23, 0.95)), url(${backgroundImage})`,
              backgroundSize: 'cover',
              backgroundPosition: 'center',
            }
          : undefined
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex items-start justify-between gap-4">
          <Text
            variant="h3"
            weight="strong"
            className="text-white group-hover:text-gradient transition-colors"
          >
            {title}
          </Text>
          <span
            className={cn(
              'px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider rounded border',
              STATUS_STYLES[status]
            )}
          >
            {status}
          </span>
        </div>
        <Text variant="body" theme="neutral" role="paragraph">
          {description}
        </Text>
      </div>
      <div className="mt-6">
        <Button
          href={href}
          variant="ghost"
          size="sm"
          className="group-hover:text-primary-400"
        >
          Try it out
          <ArrowRight
            size={14}
            weight="bold"
            className="transition-transform group-hover:translate-x-0.5"
          />
        </Button>
      </div>
    </div>
  )
}
