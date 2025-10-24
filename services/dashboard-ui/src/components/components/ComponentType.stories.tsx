import { ComponentType } from './ComponentType'

export const AllTypes = () => (
  <div className="flex flex-col gap-4">
    <div className="flex flex-wrap items-center gap-4">
      <ComponentType type="docker_build" />
      <ComponentType type="external_image" />
      <ComponentType type="helm_chart" />
      <ComponentType type="terraform_module" />
      <ComponentType type="job" />
      <ComponentType type="kubernetes_manifest" />
      <ComponentType type="unknown" />
    </div>
  </div>
)

export const DisplayVariants = () => (
  <div className="flex flex-col gap-6">
    <div>
      <h3 className="mb-2 font-medium">Full Name (default)</h3>
      <div className="flex flex-wrap items-center gap-4">
        <ComponentType type="docker_build" displayVariant="name" />
        <ComponentType type="external_image" displayVariant="name" />
        <ComponentType type="helm_chart" displayVariant="name" />
        <ComponentType type="terraform_module" displayVariant="name" />
        <ComponentType type="job" displayVariant="name" />
        <ComponentType type="kubernetes_manifest" displayVariant="name" />
      </div>
    </div>

    <div>
      <h3 className="mb-2 font-medium">Abbreviation</h3>
      <div className="flex items-center gap-4">
        <ComponentType type="docker_build" displayVariant="abbr" />
        <ComponentType type="external_image" displayVariant="abbr" />
        <ComponentType type="helm_chart" displayVariant="abbr" />
        <ComponentType type="terraform_module" displayVariant="abbr" />
        <ComponentType type="job" displayVariant="abbr" />
        <ComponentType type="kubernetes_manifest" displayVariant="abbr" />
      </div>
    </div>

    <div>
      <h3 className="mb-2 font-medium">Icon Only</h3>
      <div className="flex items-center gap-4">
        <ComponentType type="docker_build" displayVariant="icon-only" />
        <ComponentType type="external_image" displayVariant="icon-only" />
        <ComponentType type="helm_chart" displayVariant="icon-only" />
        <ComponentType type="terraform_module" displayVariant="icon-only" />
        <ComponentType type="job" displayVariant="icon-only" />
        <ComponentType type="kubernetes_manifest" displayVariant="icon-only" />
      </div>
    </div>
  </div>
)

export const TextVariants = () => (
  <div className="flex flex-col gap-4">
    <div>
      <h3 className="mb-2 font-medium">Different Text Variants</h3>
      <div className="flex flex-col gap-2">
        <ComponentType type="docker_build" variant="base" />
        <ComponentType type="docker_build" variant="subtext" />
      </div>
    </div>
  </div>
)

export const Docker = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="docker_build" displayVariant="icon-only" />
    <ComponentType type="docker_build" displayVariant="abbr" />
    <ComponentType type="docker_build" displayVariant="name" />
  </div>
)

export const ExternalImage = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="external_image" displayVariant="icon-only" />
    <ComponentType type="external_image" displayVariant="abbr" />
    <ComponentType type="external_image" displayVariant="name" />
  </div>
)

export const Helm = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="helm_chart" displayVariant="icon-only" />
    <ComponentType type="helm_chart" displayVariant="abbr" />
    <ComponentType type="helm_chart" displayVariant="name" />
  </div>
)

export const Terraform = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="terraform_module" displayVariant="icon-only" />
    <ComponentType type="terraform_module" displayVariant="abbr" />
    <ComponentType type="terraform_module" displayVariant="name" />
  </div>
)

export const Job = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="job" displayVariant="icon-only" />
    <ComponentType type="job" displayVariant="abbr" />
    <ComponentType type="job" displayVariant="name" />
  </div>
)

export const Kubernetes = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="kubernetes_manifest" displayVariant="icon-only" />
    <ComponentType type="kubernetes_manifest" displayVariant="abbr" />
    <ComponentType type="kubernetes_manifest" displayVariant="name" />
  </div>
)

export const Unknown = () => (
  <div className="flex items-center gap-4">
    <ComponentType type="unknown" displayVariant="icon-only" />
    <ComponentType type="unknown" displayVariant="abbr" />
    <ComponentType type="unknown" displayVariant="name" />
  </div>
)
