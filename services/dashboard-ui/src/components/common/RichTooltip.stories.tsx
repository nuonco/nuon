import { RichTooltip } from './RichTooltip'
import { Button } from './Button'
import { Icon } from './Icon'

export const SingleItem = () => (
  <div className="flex justify-center p-8">
    <RichTooltip
      title="Actions"
      items={[
        {
          id: 'edit',
          title: 'Edit Configuration',
          subtitle: 'Modify settings',
          leftContent: <Icon variant="Pencil" />,
          // eslint-disable-next-line
          onClick: () => console.log('Edit clicked'),
        },
      ]}
    >
      <Button>Hover for single action</Button>
    </RichTooltip>
  </div>
)

export const FourItems = () => (
  <div className="flex justify-center p-8">
    <RichTooltip
      title="Quick Actions"
      items={[
        {
          id: 'view',
          title: 'View Details',
          subtitle: 'See full information',
          leftContent: <Icon variant="Eye" />,
          href: '/details',
        },
        {
          id: 'edit',
          title: 'Edit Configuration',
          subtitle: 'Modify settings',
          leftContent: <Icon variant="Pencil" />,
          // eslint-disable-next-line
          onClick: () => console.log('Edit clicked'),
        },
        {
          id: 'duplicate',
          title: 'Duplicate Item',
          subtitle: 'Create a copy',
          leftContent: <Icon variant="Copy" />,
          // eslint-disable-next-line
          onClick: () => console.log('Duplicate clicked'),
        },
        {
          id: 'delete',
          title: 'Delete Item',
          subtitle: 'Remove permanently',
          leftContent: <Icon variant="Trash" />,
          rightContent: <Icon variant="Warning" />,
          // eslint-disable-next-line
          onClick: () => console.log('Delete clicked'),
        },
      ]}
    >
      <Button>Hover for 4 actions</Button>
    </RichTooltip>
  </div>
)

export const FifteenItems = () => (
  <div className="flex justify-center p-8">
    <RichTooltip
      title="All Components"
      position="bottom"
      maxHeight="max-h-64"
      width="w-64"
      items={[
        {
          id: 'app1',
          title: 'Frontend Application',
          subtitle: 'React web app',
          leftContent: <Icon variant="Globe" />,
          href: '/components/app1',
        },
        {
          id: 'api1',
          title: 'REST API Service',
          subtitle: 'Node.js backend',
          leftContent: <Icon variant="HardDrives" />,
          href: '/components/api1',
        },
        {
          id: 'db1',
          title: 'PostgreSQL Database',
          subtitle: 'Primary database',
          leftContent: <Icon variant="Database" />,
          href: '/components/db1',
        },
        {
          id: 'cache1',
          title: 'Redis Cache',
          subtitle: 'In-memory cache',
          leftContent: <Icon variant="Lightning" />,
          href: '/components/cache1',
        },
        {
          id: 'worker1',
          title: 'Background Workers',
          subtitle: 'Job processing',
          leftContent: <Icon variant="Cpu" />,
          href: '/components/worker1',
        },
        {
          id: 'queue1',
          title: 'Message Queue',
          subtitle: 'RabbitMQ service',
          leftContent: <Icon variant="List" />,
          href: '/components/queue1',
        },
        {
          id: 'auth1',
          title: 'Authentication Service',
          subtitle: 'OAuth provider',
          leftContent: <Icon variant="Shield" />,
          href: '/components/auth1',
        },
        {
          id: 'monitor1',
          title: 'Monitoring Stack',
          subtitle: 'Prometheus & Grafana',
          leftContent: <Icon variant="ChartBar" />,
          href: '/components/monitor1',
        },
        {
          id: 'logs1',
          title: 'Log Aggregation',
          subtitle: 'ELK stack',
          leftContent: <Icon variant="FileText" />,
          href: '/components/logs1',
        },
        {
          id: 'cdn1',
          title: 'Content Delivery Network',
          subtitle: 'CloudFront CDN',
          leftContent: <Icon variant="Cloud" />,
          href: '/components/cdn1',
        },
        {
          id: 'lb1',
          title: 'Load Balancer',
          subtitle: 'Application load balancer',
          leftContent: <Icon variant="Shuffle" />,
          href: '/components/lb1',
        },
        {
          id: 'storage1',
          title: 'Object Storage',
          subtitle: 'S3 compatible storage',
          leftContent: <Icon variant="HardDrive" />,
          href: '/components/storage1',
        },
        {
          id: 'backup1',
          title: 'Backup Service',
          subtitle: 'Automated backups',
          leftContent: <Icon variant="Archive" />,
          href: '/components/backup1',
        },
        {
          id: 'search1',
          title: 'Search Engine',
          subtitle: 'Elasticsearch cluster',
          leftContent: <Icon variant="MagnifyingGlass" />,
          href: '/components/search1',
        },
        {
          id: 'mail1',
          title: 'Email Service',
          subtitle: 'SMTP provider',
          leftContent: <Icon variant="Envelope" />,
          href: '/components/mail1',
        },
      ]}
    >
      <Button>Hover for 15 components</Button>
    </RichTooltip>
  </div>
)

export const CustomConfiguration = () => (
  <div className="flex justify-center p-8">
    <RichTooltip
      title="Custom Options"
      showCount={false}
      width="w-72"
      position="left"
      items={[
        {
          id: 'option1',
          title: 'Wide tooltip without count',
          subtitle: 'Positioned to the left',
          leftContent: <Icon variant="Faders" />,
          // eslint-disable-next-line
          onClick: () => console.log('Option 1'),
        },
        {
          id: 'option2',
          title: 'Custom right content',
          subtitle: 'With status indicator',
          leftContent: <Icon variant="CheckCircle" />,
          rightContent: <div className="w-2 h-2 bg-green-500 rounded-full" />,
          // eslint-disable-next-line
          onClick: () => console.log('Option 2'),
        },
      ]}
    >
      <Button variant="primary">Custom tooltip</Button>
    </RichTooltip>
  </div>
)

export const WithClickHandlers = () => (
  <div className="flex justify-center p-8">
    <RichTooltip
      title="Interactive Items"
      onItemClick={(item) => alert(`Clicked: ${item.title}`)}
      items={[
        {
          id: 'action1',
          title: 'Action with handler',
          subtitle: 'Triggers alert',
          leftContent: <Icon variant="Play" />,
          // eslint-disable-next-line
          onClick: () => console.log('Individual handler'),
        },
        {
          id: 'action2',
          title: 'Another action',
          subtitle: 'Also triggers alert',
          leftContent: <Icon variant="Square" />,
        },
      ]}
    >
      <Button variant="secondary">Interactive tooltip</Button>
    </RichTooltip>
  </div>
)

export const InfoOnly = () => (
  <div className="flex justify-center p-8">
    <RichTooltip
      title="System Status"
      items={[
        {
          id: 'status1',
          title: 'API Response Time',
          subtitle: '245ms average',
        },
        {
          id: 'status2',
          title: 'Database Connections',
          subtitle: '12/100 active',
        },
        {
          id: 'status3',
          title: 'Memory Usage',
          subtitle: '68% of 8GB',
        },
        {
          id: 'status4',
          title: 'CPU Utilization',
          subtitle: '23% across 4 cores',
        },
        {
          id: 'status5',
          title: 'Disk Space',
          subtitle: '156GB of 500GB used',
        },
      ]}
    >
      <Button variant="ghost">View system metrics</Button>
    </RichTooltip>
  </div>
)
