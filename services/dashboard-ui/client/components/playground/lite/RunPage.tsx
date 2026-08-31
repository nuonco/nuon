import { Outlet, useParams } from 'react-router'
import { Page } from './Page'
import { runTabs } from './nav'

export interface IRunPage {
  kind: 'action' | 'runbook' | 'deploy'
}

export const RunPage = ({ kind }: IRunPage) => {
  const {
    installId = '',
    actionId = '',
    componentId = '',
    runId = '',
  } = useParams()

  const install = `/installs/${installId}`
  const isDeploy = kind === 'deploy'
  const section = kind === 'runbook' ? 'runbooks' : 'actions'
  const sectionPath = `${install}/${section}`
  const parentId = isDeploy ? componentId : actionId
  const parentPath = isDeploy
    ? `${install}/components/${componentId}`
    : `${sectionPath}/${actionId}`
  const basePath = isDeploy
    ? `${parentPath}/deploys/${runId}`
    : `${parentPath}/runs/${runId}`

  return (
    <Page
      tabs={runTabs(basePath)}
      crumbs={[
        { label: 'Installs', path: '/installs' },
        { label: installId, path: install },
        { label: parentId, path: parentPath },
        { label: runId },
      ]}
      actions={[isDeploy ? 'Redeploy' : 'Re-run']}
    >
      <Outlet />
    </Page>
  )
}
