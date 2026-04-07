export default {
  title: 'Approvals/PlanDiffs/DiffFilter',
}

import { DiffFilter } from './DiffFilter'

export const Terraform = () => (
  <DiffFilter
    title="resources"
    diffType="terraform"
    selectedActions={new Set(['create', 'update'])}
    onInputToggle={() => {}}
    onButtonClick={() => {}}
    onReset={() => {}}
    selectedCount={3}
    totalCount={8}
    searchValue=""
    onSearchChange={() => {}}
    searchPlaceholder="Search by address, resource, or name"
  />
)

export const HelmK8s = () => (
  <DiffFilter
    title="changes"
    diffType="helm-k8s"
    selectedActions={new Set(['added', 'changed'])}
    onInputToggle={() => {}}
    onButtonClick={() => {}}
    onReset={() => {}}
    selectedCount={2}
    totalCount={5}
    searchValue=""
    onSearchChange={() => {}}
    searchPlaceholder="Search by release, resource, or type"
  />
)

export const WithSearch = () => (
  <DiffFilter
    title="drift"
    diffType="terraform"
    selectedActions={new Set(['create', 'update', 'delete', 'replace', 'read', 'no-op'])}
    onInputToggle={() => {}}
    onButtonClick={() => {}}
    onReset={() => {}}
    selectedCount={4}
    totalCount={4}
    searchValue="aws_s3_bucket"
    onSearchChange={() => {}}
    searchPlaceholder="Search by address, resource, or name"
  />
)
