import { useInstall } from '@/hooks/use-install'
import { InstallConfigGraph } from './InstallConfigGraph'

export const InstallConfigGraphContainer = () => {
  const { install } = useInstall()
  return <InstallConfigGraph appId={install.app_id} appConfigId={install.app_config_id} />
}
