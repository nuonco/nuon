'use client'

import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import { DeploySwitcher } from '@/components/deploys/DeploySwitcher'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TDeploy, TInstallComponent } from '@/types'

interface IInstallComponentHeader extends IPollingProps {
  initDeploy: TDeploy
  installComponent: TInstallComponent
}

export const InstallComponentHeader = ({
  initDeploy,
  installComponent,
  pollInterval = 20000,
  shouldPoll = false,
}: IInstallComponentHeader) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { data: deploy } = usePolling<TDeploy>({
    initData: initDeploy,
    path: `/api/orgs/${org.id}/installs/${install.id}/deploys/${initDeploy?.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <>
      <header className="flex flex-wrap items-center gap-4 justify-between w-full">
        <div className="flex flex-col gap-4">
          <BackLink />
          <HeadingGroup className="gap-1">
            <Text
              className="inline-flex items-center gap-2"
              variant="h3"
              weight="strong"
            >
              <ComponentType
                type={installComponent?.component?.type}
                displayVariant="icon-only"
              />
              {installComponent?.component?.name}
              <Status status={deploy?.status_v2?.status} variant="badge" />
            </Text>
            <ID>{deploy?.id}</ID>
            <div className="flex items-center gap-4">
              <Text
                className="flex items-center gap-1"
                variant="subtext"
                theme="info"
              >
                Deployed
                <Time
                  time={deploy?.updated_at}
                  format="relative"
                  variant="subtext"
                  theme="info"
                />
              </Text>

              <Time
                time={deploy?.updated_at}
                variant="subtext"
                theme="neutral"
              />
            </div>
          </HeadingGroup>
        </div>
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-4 md:gap-8">
            <div className="flex items-center gap-4">
              <DeploySwitcher
                componentId={installComponent.component_id}
                deployId={initDeploy?.id}
              />
              <Dropdown
                alignment="right"
                variant="primary"
                buttonText="Manage"
                id="install-component-dropdown"
              >
                <Menu className="w-56">
                  <Button>
                    Redeploy component <Icon variant="CloudArrowUp" />
                  </Button>
                  <Button>
                    Teardown component <Icon variant="CloudArrowDown" />
                  </Button>
                  {installComponent?.component?.type === 'terraform_module' ? (
                    <Button>
                      Unlock Terraform state <Icon variant="LockOpen" />
                    </Button>
                  ) : null}
                </Menu>
              </Dropdown>
            </div>
          </div>
        </div>
      </header>
    </>
  )
}
