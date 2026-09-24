import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import { PreviewBranchRunModal } from './PreviewBranchRunModal'

export default {
  title: 'Branches/PreviewBranchRunModal',
}

const branchOptions = [
  { name: 'am/add-lovable-enterprise-azure' },
  { name: 'am/gcp-custom-root-domain' },
  { name: 'am/gitops-ui' },
  { name: 'am/restore-preview-ingress' },
  { name: 'am/test-pr-ci' },
  { name: 'bakik/monitoring' },
  { name: 'docs/azure-aks-runtime-setup' },
  { name: 'fix/aws/migrate-components-out-of-stack' },
]

export const BranchSource = () => {
  const [selectedBranch, setSelectedBranch] = useState(branchOptions[0])

  return (
    <ModalStory label="Open preview run">
      <PreviewBranchRunModal
        branchName="main"
        branchDefaultsSummary="Plan only · no install · statuses off · PR comments off"
        hasOverride={false}
        sourceTab="branch"
        onSourceTabChange={() => {}}
        prOptions={[]}
        branchOptions={branchOptions}
        selectedBranch={selectedBranch}
        onSelectBranch={setSelectedBranch}
        onSelectPR={() => {}}
        installOptions={[]}
        selectedInstallId=""
        onSelectInstallId={() => {}}
        noInstallOptions
        mode="build-only"
        onModeChange={() => {}}
        loadingSources={false}
        loadingInstalls={false}
        isPending={false}
        onConfirm={() => {}}
      />
    </ModalStory>
  )
}
