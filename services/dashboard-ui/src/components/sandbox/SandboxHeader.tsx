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
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TSandboxRun, TSandboxConfig, TCloudPlatform } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { SandboxRunSwitcher } from './SandboxRunSwitcher'

interface ISandboxHeader extends IPollingProps {
  initSandboxRun: TSandboxRun
  sandboxConfig?: TSandboxConfig
}

export const SandboxHeader = ({
  initSandboxRun,
  sandboxConfig,
  pollInterval = 20000,
  shouldPoll = false,
}: ISandboxHeader) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { data: sandboxRun } = usePolling<TSandboxRun>({
    initData: initSandboxRun,
    path: `/api/orgs/${org.id}/installs/${install.id}/sandbox/runs/${initSandboxRun?.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <>
      <header className="flex flex-wrap items-center gap-4 justify-between w-full">
        <div className="flex flex-col gap-4">
          <BackLink />
          <HeadingGroup className="">
            <Text
              className="inline-flex items-center gap-4"
              variant="h3"
              weight="strong"
            >
              Sandbox {sandboxRun?.run_type}
              <Status status={sandboxRun?.status_v2?.status} variant="badge" />
            </Text>
            <Text
              className="flex items-center gap-1"
              variant="subtext"
              theme="info"
            >
              {toSentenceCase(sandboxRun?.run_type)}
              <Time
                time={sandboxRun?.updated_at}
                format="relative"
                variant="subtext"
                theme="info"
              />
            </Text>
          </HeadingGroup>
        </div>
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-4 md:gap-8">
            <div className="flex items-center gap-4">
              <SandboxRunSwitcher sandboxRunId={initSandboxRun?.id} />
              <Dropdown
                alignment="right"
                variant="primary"
                buttonText="Manage"
                id="install-component-dropdown"
              >
                <Menu className="w-56">
                  <Button>
                    Reprovision sandbox <Icon variant="CloudArrowUp" />
                  </Button>
                  <Button>
                    Deprovision sandbox <Icon variant="CloudArrowDown" />
                  </Button>

                  <Button>
                    Unlock Terraform state <Icon variant="LockOpen" />
                  </Button>
                </Menu>
              </Dropdown>
            </div>
          </div>
        </div>
      </header>
      <div className="flex items-center gap-4">
        <CloudPlatform
          platform={install.cloud_platform as TCloudPlatform}
          variant="subtext"
        />
        <Text className="flex items-center gap-2" variant="subtext">
          Install ID: <ID>{install?.id}</ID>
        </Text>
        <Text className="flex items-center gap-2" variant="subtext">
          Run ID:
          <ID>{sandboxRun?.id}</ID>
        </Text>
      </div>
    </>
  )
}
