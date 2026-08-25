import { useOutletContext } from 'react-router'
import { Markdown } from '@/components/common/Markdown'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import type { TInstallRunbookOutletContext } from './types'

export const RunbookReadmeTab = () => {
  const { installRunbook } = useOutletContext<TInstallRunbookOutletContext>()
  const { install } = useInstall()
  const latestConfig = installRunbook?.runbook?.configs?.[0]

  return (
    <>
      <PageTitle
        segments={[
          `${installRunbook?.runbook?.name ?? 'Runbook'} readme`,
          install?.name,
        ]}
      />
      {!latestConfig?.readme ? (
        <Text theme="neutral">No readme configured.</Text>
      ) : (
        <Markdown content={latestConfig.readme} mode="install" />
      )}
    </>
  )
}
