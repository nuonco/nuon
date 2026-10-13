import { Navigate } from 'react-router'
import {
  DEPLOYMENT_TABS,
  deploymentTabOrder,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { useWorkflow } from '@/hooks/use-workflow'

export const DeploymentDefaultTab = () => {
  const { workflow } = useWorkflow()
  const tab = deploymentTabOrder(workflow?.type)[0]
  return <Navigate to={DEPLOYMENT_TABS[tab].path.slice(1)} replace />
}
