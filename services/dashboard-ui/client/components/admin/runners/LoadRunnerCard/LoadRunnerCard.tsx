import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { RunnerCard } from '../RunnerCard'
import type { TRunner } from '@/types'

interface ILoadRunnerCard {
  runner?: TRunner
  error: string | null
  isLoading: boolean
  href: string
  onAction: () => void
}

export const LoadRunnerCard = ({ runner, error, isLoading, href, onAction }: ILoadRunnerCard) => {
  if (error) {
    return (
      <Text variant="subtext" className="text-red-600">
        {error}
      </Text>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 p-4">
        <Icon variant="Loading" className="animate-spin" size="16" />
        <Text variant="subtext">Loading runner...</Text>
      </div>
    )
  }

  if (!runner) {
    return <Text variant="subtext">Runner not found</Text>
  }

  return (
    <RunnerCard
      runner={runner}
      href={href}
      isInstallRunner={true}
      onAction={onAction}
    />
  )
}
