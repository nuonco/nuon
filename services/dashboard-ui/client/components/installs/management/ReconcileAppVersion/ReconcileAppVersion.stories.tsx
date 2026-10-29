export default {
  title: 'Features / Installs / Management / Reconcile app version',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ReconcileAppVersionModal } from './ReconcileAppVersion'

const noop = () => {}

export const OutOfDate = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const NeverProvisioned = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId={undefined}
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={0}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const InSync = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hinsync"
      targetConfigId="appcfg_01hinsync"
      unappliedCount={0}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithConflictError = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={false}
      error={{
        error:
          'The install is already on its pinned app config; there is nothing to reconcile.',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
