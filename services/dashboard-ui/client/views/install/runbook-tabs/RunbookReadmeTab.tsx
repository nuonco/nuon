import { useOutletContext } from 'react-router'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookReadmeTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const latestConfig = installRunbook?.runbook?.configs?.[0]

  if (!latestConfig?.readme) {
    return <Text theme="neutral">No readme configured.</Text>
  }

  return <Markdown content={latestConfig.readme} mode="install" />
}
