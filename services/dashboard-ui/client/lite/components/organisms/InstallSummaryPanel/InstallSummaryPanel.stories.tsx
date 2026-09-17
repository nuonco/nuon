import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { InstallSummaryPanel } from './InstallSummaryPanel'

export default {
  title: 'lite/organisms/InstallSummaryPanel',
}

const openPanel = (panel: React.ReactElement) => (surfaces: {
  openPanel: (content: React.ReactElement) => string
}) => {
  surfaces.openPanel(panel)
}

export const Overview = () => (
  <ComponentDocs
    name="InstallSummaryPanel"
    tier="organism"
    summary="A compact install preview with a link to the install page."
    use={[
      'Open from an install row in a deployment stage.',
      'Use the container so direct panel URLs fetch current install data.',
    ]}
    avoid={[
      'Do not reproduce the full install page in this panel.',
      'Do not add install management actions here.',
      'Do not build the install URL inside the presentation component.',
    ]}
    rules={[
      'The heading carries the install name with its ID beneath.',
      'The body is flat: platform and region, then status axes, then labels. No nested cards.',
      'The cloud platform is its brand mark alone, without the provider name.',
      'Label colours come from the owning app.',
      'View install is the single way out to full detail.',
      'Loading and failure remain inside the panel shell.',
    ]}
    props={[
      {
        name: 'install',
        type: 'TInstall',
        description: 'Install summary data.',
      },
      {
        name: 'installHref',
        type: 'string',
        description: 'Link to the full install page.',
      },
      {
        name: 'labelColors',
        type: 'Record<string, string>',
        description: 'Owning app label colours keyed by label key.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads fields within the panel shell.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Shows the install failed-to-load state.',
      },
    ]}
  />
)

export const Default = () => (
  <SurfaceStory
    open={openPanel(
      <InstallSummaryPanel
        install={{
          id: 'inst_alpha',
          name: 'alpha',
          app_id: 'app_example',
          cloud_platform: 'aws',
          aws_account: { region: 'us-west-2' },
          runner_status: 'active',
          sandbox_status: 'active',
          sandbox_health_status: 'healthy',
          composite_component_status: 'deploying',
          composite_health_status: 'healthy',
          labels: { env: 'prod', region: 'us-west-2' },
        }}
        installHref="/org_example/installs/inst_alpha"
        labelColors={{ env: '#4cc9f0', region: '#8b5cf6' }}
      />
    )}
  />
)

export const Loading = () => (
  <SurfaceStory open={openPanel(<InstallSummaryPanel loading />)} />
)

export const FailedToLoad = () => (
  <SurfaceStory
    open={openPanel(<InstallSummaryPanel error={new Error('failed')} />)}
  />
)
