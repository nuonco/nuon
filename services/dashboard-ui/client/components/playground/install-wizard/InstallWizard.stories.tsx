export default {
  title: 'Playground / Installs / Wizard',
  meta: { fullBleed: true },
}

import { InstallWizard } from './InstallWizard'

export const Create = () => <InstallWizard mode="create" />

export const Provisioning = () => <InstallWizard mode="create" startAtWorkflow />

export const Teardown = () => <InstallWizard mode="teardown" />

export const TearingDown = () => <InstallWizard mode="teardown" startAtWorkflow />
