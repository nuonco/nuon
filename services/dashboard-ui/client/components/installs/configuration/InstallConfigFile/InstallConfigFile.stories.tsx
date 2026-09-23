export default {
  title: 'Installs/Configuration/InstallConfigFile',
}

import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { InstallConfigFile } from './InstallConfigFile'

const day = 86400000
const ago = (days: number) => new Date(Date.now() - day * days).toISOString()

type TConfigEvent = { id: string; created_at: string; status: string }

const configEvents: TConfigEvent[] = [
  { id: 'icv-3', created_at: ago(0.2), status: 'active' },
  { id: 'icv-2', created_at: ago(6), status: 'error' },
  { id: 'icv-1', created_at: ago(28), status: 'active' },
]

const ConfigHistory = ({
  events = configEvents,
}: {
  events?: TConfigEvent[]
}) =>
  events.length ? (
    <Timeline<TConfigEvent>
      events={events}
      getEventKey={(event) => event.id}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      renderEvent={(event) => (
        <TimelineEvent
          createdAt={event.created_at}
          status={event.status as 'active'}
          badge={{ children: 'github', theme: 'neutral' }}
          title="Sync install config from git"
          caption={<ID>{event.id}</ID>}
          additionalCaption={
            <Badge size="sm" theme="info">
              Updated
            </Badge>
          }
        />
      )}
    />
  ) : (
    <Text variant="subtext" theme="neutral">
      Config versions appear here after the first sync from git.
    </Text>
  )

const configFileContents = `version = "v1"

[install]
name = "acme-production"
app_branch = "release"

[install.inputs]
region = "us-west-2"
cluster_size = "5"
domain = "payments.example.com"

[install.aws]
region = "us-west-2"
iam_role_arn = "arn:aws:iam::000000000000:role/nuon-install"
`

const longConfigFileContents = `version = "v1"

[install]
name = "acme-production"

[install.inputs]
${Array.from({ length: 320 }, (_, index) => `input_${index} = "value-${index}"`).join('\n')}
`

const syncAction = <Button variant="secondary">Sync now</Button>

export const Default = () => (
  <InstallConfigFile
    action={syncAction}
    content={configFileContents}
    filename="installs/acme-production.toml"
    history={<ConfigHistory />}
    isManagedByConfig
    latestVersionId="icv-3"
    syncedAt={ago(0.2)}
  />
)

export const Syncing = () => (
  <InstallConfigFile
    action={
      <Button variant="secondary" disabled>
        Syncing config
      </Button>
    }
    content={configFileContents}
    filename="installs/acme-production.toml"
    history={<ConfigHistory />}
    isManagedByConfig
    latestVersionId="icv-3"
    syncedAt={ago(0.2)}
  />
)

export const LargeFile = () => (
  <InstallConfigFile
    action={syncAction}
    content={longConfigFileContents}
    filename="installs/acme-production.toml"
    history={<ConfigHistory />}
    isManagedByConfig
    latestVersionId="icv-3"
    syncedAt={ago(0.2)}
  />
)

export const NoSyncMetadata = () => (
  <InstallConfigFile
    action={syncAction}
    content={configFileContents}
    history={<ConfigHistory events={[]} />}
    isManagedByConfig
  />
)

export const GenerateFailed = () => (
  <InstallConfigFile
    action={syncAction}
    history={<ConfigHistory />}
    isManagedByConfig
    latestVersionId="icv-3"
    syncedAt={ago(0.2)}
  />
)

export const Loading = () => (
  <InstallConfigFile
    action={syncAction}
    history={<ConfigHistory />}
    isManagedByConfig
    isLoading
  />
)

export const DashboardManaged = () => (
  <InstallConfigFile isManagedByConfig={false} />
)
