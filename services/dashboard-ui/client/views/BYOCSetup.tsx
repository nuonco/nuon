import { useQuery } from '@tanstack/react-query'
import { Avatar } from '@/components/common/Avatar'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { Logo } from '@/components/common/Logo'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { useConfig } from '@/hooks/use-config'
import { getOrgs } from '@/lib'

function Step({ number, title, children }: { number: number; title: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-3">
        <div className="w-7 h-7 rounded-full bg-primary-600 dark:bg-primary-400 flex items-center justify-center flex-shrink-0">
          <Text variant="label" weight="strong" className="text-white dark:text-black">{number}</Text>
        </div>
        <Text variant="h3">{title}</Text>
      </div>
      <div className="pl-10 flex flex-col gap-3">
        {children}
      </div>
    </div>
  )
}

export function BYOCSetup() {
  const { apiUrl } = useConfig()

  const { data: orgs } = useQuery({
    queryKey: ['byoc-setup-orgs'],
    queryFn: () => getOrgs({ limit: 1 }),
    refetchInterval: 5000,
  })

  const org = orgs?.at(0)

  const exportCmd = `export NUON_API_URL=${apiUrl}`
  const loginCmd = 'nuon auth login'
  const orgsCreateCmd = 'nuon orgs create -n "my-org"'
  const appsCreateCmd = 'nuon apps create -n "my-app"'
  const appsSyncCmd = 'nuon apps sync'

  return (
    <div className="h-screen flex flex-col bg-background">
      <div className="flex justify-between w-full px-6 pt-4 pb-4 border-b">
        <Logo />
        <Button variant="ghost" href="https://docs.nuon.co" size="sm">
          <Icon variant="BookOpenIcon" size={14} /> Docs
        </Button>
      </div>
      <div className="flex-1 overflow-y-auto px-6 pt-14 pb-8">
        <div className="max-w-2xl mx-auto w-full flex flex-col gap-10">
          <div className="flex flex-col gap-2">
            <Text variant="h1">Get started</Text>
            <Text variant="body" theme="neutral">
              Set up the Nuon CLI, create your organization, and sync your first app.
            </Text>
          </div>

          <Step number={1} title="Install the CLI">
            <Tabs
              tabs={{
                homebrew: (
                  <div className="mt-4 relative">
                    <ClickToCopyButton
                      className="w-fit !absolute right-2 top-3"
                      textToCopy="brew install nuonco/tap/nuon"
                    />
                    <CodeBlock language="bash">
                      brew install nuonco/tap/nuon
                    </CodeBlock>
                  </div>
                ),
                script: (
                  <div className="mt-4 relative">
                    <ClickToCopyButton
                      className="w-fit !absolute right-2 top-3"
                      textToCopy="bash <(curl -sSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)"
                    />
                    <CodeBlock language="bash">
                      {'bash <(curl -sSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)'}
                    </CodeBlock>
                  </div>
                ),
              }}
            />
            <a
              href="https://docs.nuon.co/cli"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary-600 dark:text-primary-400 text-sm underline underline-offset-2 w-fit"
            >
              View CLI documentation
            </a>
          </Step>

          <Step number={2} title="Log in">
            <Text variant="body" theme="neutral">
              Point the CLI at your Nuon instance and authenticate.
            </Text>
            <div className="relative">
              <ClickToCopyButton
                className="w-fit !absolute right-2 top-3"
                textToCopy={`${exportCmd}\n${loginCmd}`}
              />
              <CodeBlock language="bash">{`${exportCmd}\n${loginCmd}`}</CodeBlock>
            </div>
          </Step>

          <Step number={3} title="Create an organization">
            <Text variant="body" theme="neutral">
              Organizations are the top-level container for your apps and installs.
            </Text>
            <div className="relative">
              <ClickToCopyButton
                className="w-fit !absolute right-2 top-3"
                textToCopy={orgsCreateCmd}
              />
              <CodeBlock language="bash">{orgsCreateCmd}</CodeBlock>
            </div>
          </Step>

          <Step number={4} title="Create an app">
            <Text variant="body" theme="neutral">
              Create a new app in your organization. You can reference the{' '}
              <a
                href="https://github.com/nuonco/example-app-configs"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary-600 dark:text-primary-400 underline underline-offset-2"
              >
                example app configs
              </a>
              {' '}to get started with a working configuration.
            </Text>
            <div className="relative">
              <ClickToCopyButton
                className="w-fit !absolute right-2 top-3"
                textToCopy={appsCreateCmd}
              />
              <CodeBlock language="bash">{appsCreateCmd}</CodeBlock>
            </div>
          </Step>

          <Step number={5} title="Sync your app config">
            <Text variant="body" theme="neutral">
              From inside your app config directory, run sync to push your configuration to Nuon and trigger your first build.
              See the{' '}
              <a
                href="https://docs.nuon.co/cli-commands#apps-&-config"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary-600 dark:text-primary-400 underline underline-offset-2"
              >
                apps CLI commands
              </a>
              {' '}for more details.
            </Text>
            <div className="relative">
              <ClickToCopyButton
                className="w-fit !absolute right-2 top-3"
                textToCopy={appsSyncCmd}
              />
              <CodeBlock language="bash">{appsSyncCmd}</CodeBlock>
            </div>
          </Step>

          <div className="border-t pt-6 flex flex-col gap-4">
            {org ? (
              <div className="flex items-center gap-4 p-4 rounded-lg border bg-background-secondary">
                <Avatar
                  {...(org.logo_url ? { src: org.logo_url } : { name: org.name })}
                  size="xl"
                />
                <div className="flex-1 min-w-0">
                  <Text weight="strong" variant="subtext" flex className="text-nowrap">
                    {org.sandbox_mode && (
                      <Icon variant="TestTubeIcon" className="!w-[12px] !h-[12px] shrink-0" size="12" />
                    )}
                    <span className="truncate">{org.name}</span>
                  </Text>
                  <Status status={org.status} />
                </div>
                <Button variant="primary" href={`/${org.id}`}>
                  Go to dashboard <Icon variant="ArrowRightIcon" weight="bold" />
                </Button>
              </div>
            ) : (
              <div className="flex items-center gap-3 p-4 rounded-lg border">
                <Icon variant="Loading" size={16} />
                <Text variant="body" theme="neutral">
                  Waiting for an organization to be created...
                </Text>
              </div>
            )}
            <a
              href="https://docs.nuon.co/get-started/create-your-first-app"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary-600 dark:text-primary-400 text-sm underline underline-offset-2 w-fit"
            >
              Create your first app guide
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}
