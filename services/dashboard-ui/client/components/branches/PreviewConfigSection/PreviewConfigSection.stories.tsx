import type { ReactNode } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import type {
  TAPIError,
  TAppBranchConfig,
  TInstall,
} from '@/types'
import { PreviewConfigEditorModal } from './PreviewConfigEditorModal'
import { PreviewConfigSection } from './PreviewConfigSection'

export default {
  title: 'Branches/PreviewConfigSection',
}

const installs = [
  { id: 'ins_acme_dev', name: 'acme-dev' },
  { id: 'ins_acme_stage', name: 'acme-stage' },
] as TInstall[]

const config = {
  config_number: 12,
  preview_config: {
    mode: 'plan-only',
    install_id: 'ins_acme_dev',
    set_statuses: true,
    comment: false,
  },
} as TAppBranchConfig

const StoryFrame = ({ children }: { children: ReactNode }) => (
  <div className="max-w-3xl">{children}</div>
)

export const Default = () => (
  <StoryFrame>
    <PreviewConfigSection
      currentConfig={config}
      installs={installs}
      hasGithubVCS
    />
  </StoryFrame>
)

export const PlatformDefaults = () => (
  <StoryFrame>
    <PreviewConfigSection installs={installs} hasGithubVCS />
  </StoryFrame>
)

export const BuildOnly = () => (
  <StoryFrame>
    <PreviewConfigSection
      currentConfig={
        {
          config_number: 13,
          preview_config: {
            mode: 'build-only',
            set_statuses: true,
            comment: true,
          },
        } as TAppBranchConfig
      }
      installs={installs}
      hasGithubVCS
    />
  </StoryFrame>
)

export const LabelSelector = () => (
  <StoryFrame>
    <PreviewConfigSection
      currentConfig={
        {
          config_number: 14,
          preview_config: {
            mode: 'apply',
            label_selector: {
              match_labels: {
                environment: 'staging',
                region: 'us-west-2',
              },
            },
            set_statuses: false,
            comment: false,
          },
        } as TAppBranchConfig
      }
      installs={installs}
      hasGithubVCS
    />
  </StoryFrame>
)

export const Loading = () => (
  <StoryFrame>
    <PreviewConfigSection
      currentConfig={config}
      installs={installs}
      hasGithubVCS
      isLoading
    />
  </StoryFrame>
)

export const EditModal = () => (
  <ModalStory label="Edit preview settings">
    <PreviewConfigEditorModal
      currentConfig={config}
      installs={installs}
      hasGithubVCS
      onSubmit={() => undefined}
      onCancel={() => undefined}
    />
  </ModalStory>
)

export const EditModalError = () => (
  <ModalStory label="Open error state">
    <PreviewConfigEditorModal
      currentConfig={config}
      installs={installs}
      hasGithubVCS
      error={
        {
          error: 'Preview settings save failed',
          description: 'The config changed before these settings were saved.',
        } as TAPIError
      }
      onSubmit={() => undefined}
      onCancel={() => undefined}
    />
  </ModalStory>
)
