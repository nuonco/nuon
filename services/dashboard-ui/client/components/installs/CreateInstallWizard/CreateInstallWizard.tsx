import {
  InstallWizard,
  type IInstallWizard,
} from '@/components/installs/InstallWizard'

export type ICreateInstallWizard = IInstallWizard

export const CreateInstallWizard = (props: ICreateInstallWizard) => (
  <InstallWizard {...props} />
)
