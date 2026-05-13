import { Icon, type TIconVariant } from './Icon'

export default { title: 'Common/Icon' }

const customVariants: TIconVariant[] = [
  'AWS',
  'AWSColor',
  'AWSLambda',
  'Azure',
  'AzureColor',
  'Docker',
  'GCP',
  'GCPColor',
  'GitHub',
  'Helm',
  'Kubernetes',
  'Loading',
  'OCI',
  'Pulumi',
  'Terraform',
]

const sampledPhosphorVariants: TIconVariant[] = [
  'ArrowLeftIcon',
  'ArrowRightIcon',
  'ArrowUpIcon',
  'ArrowDownIcon',
  'CheckIcon',
  'XIcon',
  'PlusIcon',
  'MinusIcon',
  'MagnifyingGlassIcon',
  'FunnelIcon',
  'BellIcon',
  'GearIcon',
  'TrashIcon',
  'PencilIcon',
  'CopyIcon',
  'LinkIcon',
  'WarningIcon',
  'InfoIcon',
  'CheckCircleIcon',
  'XCircleIcon',
  'SparkleIcon',
  'HouseIcon',
  'UserIcon',
  'UsersIcon',
  'LockIcon',
  'KeyIcon',
  'CloudIcon',
  'DatabaseIcon',
  'TerminalIcon',
  'CodeBlockIcon',
  'FileTextIcon',
  'FolderOpenIcon',
  'ChartBarIcon',
  'ClockIcon',
  'CalendarBlankIcon',
  'EnvelopeIcon',
  'EyeIcon',
  'EyeSlashIcon',
  'DotsThreeVerticalIcon',
  'CaretDownIcon',
  'CaretRightIcon',
]

const Row = ({ variant }: { variant: TIconVariant }) => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0' }}>
    <Icon variant={variant} size={20} />
    <span style={{ fontFamily: 'monospace', fontSize: 13 }}>{String(variant)}</span>
  </div>
)

export const CustomIcons = () => (
  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 4 }}>
    {customVariants.map((v) => (
      <Row key={String(v)} variant={v} />
    ))}
  </div>
)

export const PhosphorSamples = () => (
  <div>
    <p style={{ fontFamily: 'sans-serif', fontSize: 13, marginBottom: 12, color: '#888' }}>
      A curated sample. For the full set see{' '}
      <a href="https://phosphoricons.com" target="_blank" rel="noreferrer">
        phosphoricons.com
      </a>
      {' '}— append <code>Icon</code> to the name when using as a variant (e.g. <code>HouseIcon</code>).
    </p>
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 4 }}>
      {sampledPhosphorVariants.map((v) => (
        <Row key={String(v)} variant={v} />
      ))}
    </div>
  </div>
)

export const Themes = () => (
  <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
    {(['default', 'neutral', 'info', 'warn', 'error', 'success', 'brand'] as const).map((theme) => (
      <div key={theme} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <Icon variant="BellIcon" size={20} theme={theme} />
        <span style={{ fontFamily: 'monospace', fontSize: 13 }}>{theme}</span>
      </div>
    ))}
  </div>
)

export const Sizes = () => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
    {[12, 16, 20, 24, 32, 48].map((size) => (
      <div key={size} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
        <Icon variant="HouseIcon" size={size} />
        <span style={{ fontFamily: 'monospace', fontSize: 11 }}>{size}</span>
      </div>
    ))}
  </div>
)
