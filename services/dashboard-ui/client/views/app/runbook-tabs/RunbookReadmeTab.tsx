import { useOutletContext } from 'react-router'
import { Markdown } from '@/components/common/Markdown'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import type { TRunbookOutletContext } from './types'

export const RunbookReadmeTab = () => {
  const { runbook } = useOutletContext<TRunbookOutletContext>()
  const { app } = useApp()
  const latestConfig = runbook?.configs?.[0]

  return (
    <>
      <PageTitle
        segments={[`${runbook?.name ?? 'Runbook'} readme`, app?.name]}
      />
      {!latestConfig?.readme ? (
        <Text theme="neutral">No readme configured.</Text>
      ) : (
        <Markdown content={latestConfig.readme} mode="app" />
      )}
    </>
  )
}
