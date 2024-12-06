import classNames from 'classnames'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ActionTriggerType,
  Button,
  InstallPageSubNav,
  InstallStatuses,
  DashboardContent,
  Link,
  Section,
  Text,
  Time,
} from '@/components'
import {
  getAppWorkflows,
  getInstall,
  getInstallWorkflowRuns,
  getOrg,
} from '@/lib'

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, install, workflows] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
    getInstallWorkflowRuns({ installId, orgId }),
  ])

  const appWorkflows = await getAppWorkflows({ appId: install.app_id, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        { href: `/${org.id}/installs/${install.id}`, text: install.name },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Workflow history">
          <div className="flex flex-col gap-2">
            {workflows.map((w, i) => (
              <Link
                key={w.id}
                className="!block w-full !p-0"
                href={`/${orgId}/installs/${installId}/workflows/${w.id}`}
                variant="ghost"
              >
                <div
                  className={classNames(
                    'flex items-center justify-between p-4',
                    {
                      'border rounded-md shadow-sm': i === 0,
                    }
                  )}
                >
                  <div className="flex flex-col">
                    <span className="flex items-center gap-2">
                      <span
                        className={classNames('w-1.5 h-1.5 rounded-full', {
                          'bg-green-800 dark:bg-green-500': true,
                        })}
                      />
                      <Text variant="med-12">Succeeded</Text>
                    </span>

                    <Text
                      className="flex items-center gap-2 ml-3.5"
                      variant="reg-12"
                    >
                      <span>
                        {
                          appWorkflows?.find(
                            (aw) => aw.id === w.action_workflow_config_id
                          )?.name
                        }
                      </span>{' '}
                      /
                      <span className="!inline truncate max-w-[100px]">
                        {w._.triggers[0].type}
                      </span>
                    </Text>
                  </div>

                  <div className="flex items-center gap-2">
                    <Time
                      time={w.updated_at}
                      format="relative"
                      variant="reg-12"
                    />
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Action workflows">
            <div className="flex flex-col gap-2 divide-y">
              {appWorkflows.map((aW) => (
                <div
                  key={aW.id}
                  className="flex items-end justify-between py-2 flex-wrap gap-2"
                >
                  <div key={aW.id} className="flex flex-col gap-2">
                    <Link
                      href={`/${orgId}/apps/${install.app_id}/workflows/${aW.id}`}
                    >
                      <Text variant="med-12">{aW.name}</Text>
                    </Link>
                    <span className="flex items-center justify-start gap-2 flex-wrap">
                      {aW.configs[0].triggers?.map((t) => (
                        <ActionTriggerType key={t.id} triggerType={t.type} />
                      ))}
                      <Text variant="reg-12">
                        {aW.configs[0].steps.length} Steps
                      </Text>
                    </span>
                  </div>
                  {aW.configs[0].triggers.find((t) => t.type === 'manual') ? (
                    <Button className="text-sm !py-2 !h-fit">
                      Run workflow
                    </Button>
                  ) : null}
                </div>
              ))}
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
