import { Button } from './Button'
import { Icon } from './Icon'
import { Status } from './Status'
import { Text } from './Text'
import { Tooltip } from './Tooltip'

export const Default = () => (
  <div className="flex p-8">
    <Tooltip tipContent="This is a tooltip">
      <Text>Hover me</Text>
    </Tooltip>
  </div>
)

export const Positions = () => (
  <div className="flex gap-8 p-8">
    <Tooltip tipContent="Top" position="top">
      <Text>Top</Text>
    </Tooltip>
    <Tooltip tipContent="Bottom" position="bottom">
      <Text>Bottom</Text>
    </Tooltip>
    <Tooltip tipContent="Left" position="left">
      <Text>Left</Text>
    </Tooltip>
    <Tooltip tipContent="Right" position="right">
      <Text>Right</Text>
    </Tooltip>
  </div>
)

export const WithIcon = () => (
  <div className="p-8">
    <Tooltip tipContent="This is a tooltip" showIcon>
      <Text>Hover me</Text>
    </Tooltip>
  </div>
)

const deps = [
  {
    status: 'healthy',
    name: 'ctl_api_something-long-name-dude',
  },
  {
    status: 'deprovisioned',
    name: 'dashboard_ui',
  },
  {
    status: 'error',
    name: 'coder_db',
  },
  {
    status: 'in-progress',
    name: 'httpbin',
  },
  {
    status: 'healthy',
    name: 'auth_service',
  },
  {
    status: 'healthy',
    name: 'user_mgmt',
  },
  {
    status: 'error',
    name: 'redis_cache',
  },
  {
    status: 'in-progress',
    name: 'file_storage',
  },
  {
    status: 'healthy',
    name: 'notification_service',
  },
  {
    status: 'deprovisioned',
    name: 'legacy_api',
  },
  {
    status: 'healthy',
    name: 'workspace_manager',
  },
  {
    status: 'error',
    name: 'metrics_collector',
  },
  {
    status: 'in-progress',
    name: 'backup_service',
  },
  {
    status: 'healthy',
    name: 'load_balancer',
  },
  {
    status: 'healthy',
    name: 'cdn_proxy',
  },
  {
    status: 'error',
    name: 'log_aggregator',
  },
  {
    status: 'in-progress',
    name: 'image_registry',
  },
  {
    status: 'healthy',
    name: 'secret_manager',
  },
  {
    status: 'deprovisioned',
    name: 'old_dashboard',
  },
  {
    status: 'healthy',
    name: 'health_checker',
  },
]

export const RichTooltip = () => (
  <div className="flex p-8">
    <Tooltip
      isOpen
      position="right"
      tipContentClassName="!p-0"
      tipContent={
        <div className="flex flex-col w-52">
          <Text className="px-3 py-2 border-b flex items-cetner justify-between !leading-none">
            Dependencies
            <span>{deps.length}</span>
          </Text>
          <div className="flex flex-col divide-y max-h-30 overflow-y-auto">
            {deps.map((d) => (
              <ComponentDepButton key={d.name} {...d} />
            ))}
          </div>
        </div>
      }
    >
      <Text>Dependencies</Text>
    </Tooltip>
  </div>
)

const ComponentDepButton = ({ name, status }) => (
  <div className="shrink-0 grow-0 last-of-type:rounded-b-lg overflow-hidden">
    <Button
      className="flex items-center justify-between gap-2 w-full !rounded-none !px-3 !py-2"
      variant="ghost"
    >
      <span className="flex flex-col text-left">
        <Text
          className="flex items-center gap-2"
          variant="label"
          weight="strong"
        >
          <Status status={status} isWithoutText />
          <span className="max-w-36 truncate">{name}</span>
        </Text>
        <Text className="ml-3.5" variant="label" theme="neutral">
          {status}
        </Text>
      </span>

      <Icon variant="CaretRight" />
    </Button>
  </div>
)
