'use client'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

export function RefreshLogStream() {
  const handleRefresh = () => {
    window.location.reload()
  }

  return (
    <div className="flex flex-col items-center gap-4 p-8">
      <Text variant="base" weight="strong">
        Waiting on log stream
      </Text>
      <Button onClick={handleRefresh}>
        <Icon variant="ArrowClockwise" />
        Refresh Page
      </Button>
    </div>
  )
}
