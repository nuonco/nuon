import type { TCloudPlatform } from '@/types'
import { Badge } from '../../atoms/Badge'
import { Text } from '../../atoms/Text'
import { AppSource } from '../../molecules/AppSource'
import { CloudPlatform } from '../../molecules/CloudPlatform'
import { FormSelect, type IFormSelect } from '../../molecules/FormSelect'

export type TAppSelectReadiness =
  | 'not-provisionable'
  | 'no-valid-config'
  | 'no-components'
  | 'no-component-builds'

export const APP_SELECT_READINESS_LABEL: Record<TAppSelectReadiness, string> = {
  'not-provisionable': 'Not provisionable',
  'no-valid-config': 'No valid config',
  'no-components': 'No components',
  'no-component-builds': 'No component builds',
}

export interface IAppSelectItem {
  id: string
  name: string
  platform: TCloudPlatform
  source?: string
  updatedLabel: string
  readiness?: TAppSelectReadiness
}

export interface IAppSelect
  extends Omit<
    IFormSelect,
    | 'options'
    | 'label'
    | 'placeholder'
    | 'searchable'
    | 'searchPlaceholder'
    | 'emptyMessage'
  > {
  apps: IAppSelectItem[]
}

const Dot = () => (
  <Text variant="caption" color="tertiary" aria-hidden>
    ·
  </Text>
)

const AppName = ({ app }: { app: IAppSelectItem }) => (
  <span className="flex min-w-0 items-center gap-2">
    <span className="truncate">{app.name}</span>
    {app.readiness ? (
      <Badge className="shrink-0">
        {APP_SELECT_READINESS_LABEL[app.readiness]}
      </Badge>
    ) : null}
  </span>
)

const AppDetails = ({ app }: { app: IAppSelectItem }) => (
  <span className="flex min-w-0 items-center gap-1.5">
    <CloudPlatform
      platform={app.platform}
      display="icon"
      tooltip={false}
      iconSize={18}
      className="shrink-0"
    />
    {app.source ? (
      <>
        <Dot />
        <AppSource
          source={app.source}
          iconSize={14}
          color="tertiary"
          className="truncate"
        />
      </>
    ) : null}
    <Dot />
    <span className="shrink-0">{app.updatedLabel}</span>
  </span>
)

export const AppSelect = ({ apps, ...props }: IAppSelect) => (
  <FormSelect
    {...props}
    label="App"
    options={apps.map((app) => ({
      value: app.id,
      label: <AppName app={app} />,
      textValue: [app.name, app.id, app.source].filter(Boolean).join(' '),
      description: <AppDetails app={app} />,
      disabled:
        app.readiness === 'not-provisionable' ||
        app.readiness === 'no-valid-config',
    }))}
    placeholder="Choose an app"
    searchable
    searchPlaceholder="Search apps"
    emptyMessage="No apps found"
  />
)
