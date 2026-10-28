export default {
  title: 'Playground / Customer portal / Customer install shell',
  meta: { fullBleed: true },
}

import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { SharedInstallsStoryChrome } from '../story-chrome'
import { CustomerInstallShell } from './CustomerInstallShell'

const basePath = '/org-1/shared-installs/inst-1'

export const AllScopes = () => (
  <SharedInstallsStoryChrome>
    <CustomerInstallShell
      installName="Acme production"
      installId="instxxxxxxxxxxxxxxxxxxxxxxxx"
      basePath={basePath}
      scopes={['approvals', 'readme', 'resources', 'operations']}
    >
      <PageSection>
        <Text>All shared scopes are visible in the subnav.</Text>
      </PageSection>
    </CustomerInstallShell>
  </SharedInstallsStoryChrome>
)

export const ReadmeAndResources = () => (
  <SharedInstallsStoryChrome>
    <CustomerInstallShell
      installName="Acme production"
      installId="instxxxxxxxxxxxxxxxxxxxxxxxx"
      basePath={basePath}
      scopes={['readme', 'resources']}
    >
      <PageSection>
        <Text>Only README and resources were shared.</Text>
      </PageSection>
    </CustomerInstallShell>
  </SharedInstallsStoryChrome>
)

export const BaselineOnly = () => (
  <SharedInstallsStoryChrome>
    <CustomerInstallShell
      installName="Acme production"
      installId="instxxxxxxxxxxxxxxxxxxxxxxxx"
      basePath={basePath}
      scopes={[]}
    >
      <PageSection>
        <Text>Baseline overview and deployments only.</Text>
      </PageSection>
    </CustomerInstallShell>
  </SharedInstallsStoryChrome>
)
