import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'

export const DeployLogsTab = () => {
  const { deploy } = useDeploy()
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Deploy logs', install?.name]} />
      <LogsPanel logStream={deploy?.log_stream} />
    </>
  )
}
