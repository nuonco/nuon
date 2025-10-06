import { CloudPlatform } from './CloudPlatform'

export const AllPlatforms = () => (
  <div className="flex flex-col gap-4">
    <div className="flex items-center gap-4">
      <CloudPlatform platform="aws" />
      <CloudPlatform platform="azure" />
      <CloudPlatform platform="gcp" />
      <CloudPlatform platform="unknown" />
    </div>
  </div>
)

export const DisplayVariants = () => (
  <div className="flex flex-col gap-6">
    <div>
      <h3 className="mb-2 font-medium">Abbreviation (default)</h3>
      <div className="flex items-center gap-4">
        <CloudPlatform platform="aws" displayVariant="abbr" />
        <CloudPlatform platform="azure" displayVariant="abbr" />
        <CloudPlatform platform="gcp" displayVariant="abbr" />
        <CloudPlatform platform="unknown" displayVariant="abbr" />
      </div>
    </div>

    <div>
      <h3 className="mb-2 font-medium">Full Name</h3>
      <div className="flex items-center gap-4">
        <CloudPlatform platform="aws" displayVariant="name" />
        <CloudPlatform platform="azure" displayVariant="name" />
        <CloudPlatform platform="gcp" displayVariant="name" />
        <CloudPlatform platform="unknown" displayVariant="name" />
      </div>
    </div>

    <div>
      <h3 className="mb-2 font-medium">Icon Only</h3>
      <div className="flex items-center gap-4">
        <CloudPlatform platform="aws" displayVariant="icon-only" />
        <CloudPlatform platform="azure" displayVariant="icon-only" />
        <CloudPlatform platform="gcp" displayVariant="icon-only" />
        <CloudPlatform platform="unknown" displayVariant="icon-only" />
      </div>
    </div>
  </div>
)

export const TextVariants = () => (
  <div className="flex flex-col gap-4">
    <div>
      <h3 className="mb-2 font-medium">Different Text Variants</h3>
      <div className="flex flex-col gap-2">
        <CloudPlatform platform="aws" variant="base" />
        <CloudPlatform platform="aws" variant="subtext" />
      </div>
    </div>
  </div>
)

export const AWS = () => (
  <div className="flex items-center gap-4">
    <CloudPlatform platform="aws" displayVariant="icon-only" />
    <CloudPlatform platform="aws" displayVariant="abbr" />
    <CloudPlatform platform="aws" displayVariant="name" />
  </div>
)

export const Azure = () => (
  <div className="flex items-center gap-4">
    <CloudPlatform platform="azure" displayVariant="icon-only" />
    <CloudPlatform platform="azure" displayVariant="abbr" />
    <CloudPlatform platform="azure" displayVariant="name" />
  </div>
)

export const GCP = () => (
  <div className="flex items-center gap-4">
    <CloudPlatform platform="gcp" displayVariant="icon-only" />
    <CloudPlatform platform="gcp" displayVariant="abbr" />
    <CloudPlatform platform="gcp" displayVariant="name" />
  </div>
)

export const Unknown = () => (
  <div className="flex items-center gap-4">
    <CloudPlatform platform="unknown" displayVariant="icon-only" />
    <CloudPlatform platform="unknown" displayVariant="abbr" />
    <CloudPlatform platform="unknown" displayVariant="name" />
  </div>
)
