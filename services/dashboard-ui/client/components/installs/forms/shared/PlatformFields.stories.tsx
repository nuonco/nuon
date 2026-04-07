export default {
  title: 'Installs/Forms/PlatformFields',
}

import { PlatformFields } from './PlatformFields'

export const AWS = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="aws" />
  </form>
)

export const Azure = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="azure" />
  </form>
)

export const AWSwithDraft = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="aws" draftValues={{ region: 'us-west-2' }} />
  </form>
)
