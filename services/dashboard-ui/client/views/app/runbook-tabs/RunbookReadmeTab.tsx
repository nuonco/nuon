import { useOutletContext } from 'react-router'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import type { TRunbookOutletContext } from './types'

export const RunbookReadmeTab = () => {
  const { runbook } = useOutletContext<TRunbookOutletContext>()
  const latestConfig = runbook?.configs?.[0]

  if (!latestConfig?.readme) {
    return <Text theme="neutral">No readme configured.</Text>
  }

  return <Markdown content={latestConfig.readme} mode="app" />
}
