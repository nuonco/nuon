import type { TApp } from '@/types'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { AppSource, appSourceFromApp } from './AppSource'

export default {
  title: 'lite/molecules/AppSource',
}

export const Overview = () => (
  <ComponentDocs
    name="AppSource"
    tier="molecule"
    summary="A compact Git repository identity for an app's source."
    use={[
      'Use wherever an app source repository appears in Lite.',
      'Use appSourceFromApp when resolving the source from an API app.',
      'Pass href when the repository should navigate to its external source.',
    ]}
    avoid={[
      'Do not repeat the GitHub mark beside this component.',
      'Do not manually resolve the app source fields at each call site.',
    ]}
    rules={[
      'The custom GitHub mark and mono repository name are always paired.',
      'Long repository names truncate within their container.',
      'A missing source renders the supplied fallback without a GitHub mark.',
      'External links use the shared Link behavior.',
    ]}
    props={[
      {
        name: 'source',
        type: 'string | null',
        description: 'Repository owner and name.',
      },
      {
        name: 'href',
        type: 'string',
        description: 'Optional external repository URL.',
      },
      {
        name: 'variant',
        type: 'TTextVariant',
        default: "'caption'",
        description: 'Text size for the repository name.',
      },
      {
        name: 'color',
        type: 'TTextColor',
        default: "'secondary'",
        description: 'Text color when the source is not a link.',
      },
      {
        name: 'iconSize',
        type: 'number',
        default: '16',
        description: 'GitHub mark size in pixels.',
      },
      {
        name: 'fallback',
        type: 'ReactNode',
        default: "'—'",
        description: 'Content shown when no source is available.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Keeps the mark visible and loads the repository name.',
      },
      {
        name: 'loadingWidth',
        type: 'number',
        default: '18',
        description: 'Loading text width in characters.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-sm p-8">
    <AppSource source="acme/payments" />
  </div>
)

export const ExternalLink = () => (
  <div className="max-w-sm p-8">
    <AppSource source="acme/payments" href="https://github.com/acme/payments" />
  </div>
)

export const Missing = () => (
  <div className="max-w-sm p-8">
    <AppSource source={undefined} fallback="No source configured" />
  </div>
)

export const Loading = () => (
  <div className="max-w-sm p-8">
    <AppSource loading />
  </div>
)

export const LongSource = () => (
  <div className="w-56 p-8">
    <AppSource source="acme/payments-platform-consolidated-ledger-service" />
  </div>
)

export const ResolvedFromApp = () => {
  const app: TApp = {
    config_repo: 'acme/payments',
    sandbox_config: {
      connected_github_vcs_config: {
        repo: 'acme/payments-sandbox',
      },
    },
  }

  return (
    <div className="max-w-sm p-8">
      <AppSource source={appSourceFromApp(app)} />
    </div>
  )
}
