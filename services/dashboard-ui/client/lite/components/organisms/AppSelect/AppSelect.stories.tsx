import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { AppSelect, type IAppSelect, type IAppSelectItem } from './AppSelect'

export default {
  title: 'lite/organisms/AppSelect',
}

const APPS: IAppSelectItem[] = [
  {
    id: 'app98e2wpzdxwoey393edtqj45',
    name: 'Payments API',
    platform: 'aws',
    source: 'acme/payments',
    updatedLabel: 'synced 8 minutes ago',
  },
  {
    id: 'app7fplr1up5atx5zpxotbabm',
    name: 'Support portal',
    platform: 'aws',
    source: 'acme/support-portal',
    updatedLabel: 'synced yesterday',
  },
  {
    id: 'appk933tcyzji01s7us3aeo3x',
    name: 'Analytics warehouse',
    platform: 'azure',
    source: 'acme/analytics',
    updatedLabel: 'synced 3 days ago',
  },
  {
    id: 'appm41s7us3aeo3xk933tcyz',
    name: 'Realtime ingest',
    platform: 'gcp',
    source: 'acme/realtime-ingest',
    updatedLabel: 'synced 25 minutes ago',
  },
]

export const Overview = () => (
  <ComponentDocs
    name="AppSelect"
    tier="organism"
    summary="A form-aware Select for choosing which app an install belongs to."
    use={[
      'Use in install setup to pick the app the install is created from.',
      'Pass normalized app details from the data-fetching container.',
    ]}
    avoid={[
      'Do not name a branch or a config version in an option, because an app has many of both.',
      'Do not fetch apps inside the presentational component.',
      'Do not disable an option by hand; pass its readiness instead.',
    ]}
    rules={[
      'The app name is the option, with the platform icon, source repo and sync recency beneath it.',
      'The platform is the icon alone, with its name kept for screen readers.',
      'Search matches the app name, its id, or its source repo.',
      'A readiness value adds a badge that stays visible on the trigger once selected.',
      'Only not-provisionable disables an option, because the other states are still installable.',
      'Search and empty-state copy are owned by the component.',
    ]}
    props={[
      {
        name: 'apps',
        type: 'IAppSelectItem[]',
        description: 'Normalized apps available for selection.',
      },
      {
        name: 'field',
        type: 'AnyFieldApi',
        description: 'TanStack Form field storing the selected app id.',
      },
      {
        name: 'disabled',
        type: 'boolean',
        default: 'false',
        description: 'Makes the whole select unavailable.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the field loading state.',
      },
    ]}
  />
)

const AppSelectStory = ({
  initialValue = '',
  apps = APPS,
  ...props
}: Omit<IAppSelect, 'field' | 'apps'> & {
  initialValue?: string
  apps?: IAppSelectItem[]
}) => {
  const form = useForm({ defaultValues: { appId: initialValue } })

  return (
    <form
      className="max-w-md p-8"
      autoComplete="off"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field
        name="appId"
        validators={{ onBlur: z.string().min(1, 'Choose an app') }}
      >
        {(field) => <AppSelect field={field} apps={apps} {...props} />}
      </form.Field>
    </form>
  )
}

export const Default = () => <AppSelectStory />

export const Selected = () => (
  <AppSelectStory initialValue="app7fplr1up5atx5zpxotbabm" />
)

export const Readiness = () => (
  <AppSelectStory
    apps={[
      ...APPS,
      {
        id: 'appq7fplr1up5atx5zpxotba2',
        name: 'Legacy worker',
        platform: 'aws',
        source: 'acme/legacy-worker',
        updatedLabel: 'synced 2 weeks ago',
        readiness: 'not-provisionable',
      },
      {
        id: 'appr7fplr1up5atx5zpxotba3',
        name: 'Edge gateway',
        platform: 'gcp',
        source: 'acme/edge-gateway',
        updatedLabel: 'synced 1 hour ago',
        readiness: 'no-components',
      },
      {
        id: 'apps7fplr1up5atx5zpxotba4',
        name: 'Billing jobs',
        platform: 'aws',
        source: 'acme/billing-jobs',
        updatedLabel: 'synced 4 hours ago',
        readiness: 'no-component-builds',
      },
    ]}
  />
)

export const ReadinessSelected = () => (
  <AppSelectStory
    initialValue="appr7fplr1up5atx5zpxotba3"
    apps={[
      APPS[0],
      {
        id: 'appr7fplr1up5atx5zpxotba3',
        name: 'Edge gateway',
        platform: 'gcp',
        source: 'acme/edge-gateway',
        updatedLabel: 'synced 1 hour ago',
        readiness: 'no-components',
      },
    ]}
  />
)

export const NoSource = () => (
  <AppSelectStory apps={APPS.map(({ source: _source, ...app }) => app)} />
)

export const UnknownPlatform = () => (
  <AppSelectStory
    apps={[
      {
        id: 'appt7fplr1up5atx5zpxotba5',
        name: 'Unconfigured app',
        platform: 'unknown',
        updatedLabel: 'never synced',
      },
    ]}
  />
)

export const LongNames = () => (
  <AppSelectStory
    initialValue="appu7fplr1up5atx5zpxotba6"
    apps={[
      {
        id: 'appu7fplr1up5atx5zpxotba6',
        name: 'Payments platform consolidated ledger and reconciliation service',
        platform: 'aws',
        source: 'acme/payments-platform-consolidated-ledger-service',
        updatedLabel: 'synced 8 minutes ago',
        readiness: 'no-components',
      },
      ...APPS,
    ]}
  />
)

export const Disabled = () => (
  <AppSelectStory initialValue="app98e2wpzdxwoey393edtqj45" disabled />
)

export const Loading = () => <AppSelectStory loading />

export const Empty = () => <AppSelectStory apps={[]} />
