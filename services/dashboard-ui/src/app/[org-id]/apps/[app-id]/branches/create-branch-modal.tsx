'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { ModalBase } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import type { TVCSConnection } from '@/types'
import { createAppBranch } from './create-branch-action'

interface ICreateBranchModal {
  appId: string
  orgId: string
  isOpen: boolean
  onClose: () => void
}

export const CreateBranchModal = ({
  appId,
  orgId,
  isOpen,
  onClose,
}: ICreateBranchModal) => {
  const router = useRouter()
  const { org } = useOrg()
  const [name, setName] = useState('')
  const [vcsConnectionId, setVcsConnectionId] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Get VCS connections from org data
  const vcsConnections = org?.vcs_connections || []

  useEffect(() => {
    if (isOpen && vcsConnections.length > 0 && !vcsConnectionId) {
      // Set first VCS connection as default when modal opens
      if (vcsConnections[0]?.id) {
        setVcsConnectionId(vcsConnections[0].id)
      }
    }
  }, [isOpen, vcsConnections, vcsConnectionId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    if (!name.trim()) {
      setError('Branch name is required')
      setLoading(false)
      return
    }

    if (!vcsConnectionId) {
      setError('VCS connection is required')
      setLoading(false)
      return
    }

    const result = await createAppBranch(orgId, appId, {
      name: name.trim(),
      vcs_connection_id: vcsConnectionId,
    })

    if (result.success && result.branch) {
      setName('')
      setVcsConnectionId('')
      onClose()
      router.push(`/${orgId}/apps/${appId}/branches/${result.branch.id}`)
      router.refresh()
    } else {
      setError(result.error || 'Failed to create branch')
    }

    setLoading(false)
  }

  const handleClose = () => {
    setName('')
    setVcsConnectionId('')
    setError(null)
    onClose()
  }

  return (
    <ModalBase
      isVisible={isOpen}
      onClose={handleClose}
      heading="Create App Branch"
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-4 p-6">
        {error && <Banner theme="error">{error}</Banner>}

        {vcsConnections.length === 0 ? (
          <Banner theme="warning">
            No VCS connections found. Please connect your GitHub account first.
          </Banner>
        ) : (
          <>
            <div className="flex flex-col gap-2">
              <label htmlFor="branch-name" className="text-sm font-medium">
                Branch Name
              </label>
              <Input
                id="branch-name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="main"
                required
                disabled={loading}
              />
            </div>

            <div className="flex flex-col gap-2">
              <label htmlFor="vcs-connection" className="text-sm font-medium">
                VCS Connection
              </label>
              <select
                id="vcs-connection"
                value={vcsConnectionId}
                onChange={(e) => setVcsConnectionId(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
                disabled={loading}
              >
                {vcsConnections.map((conn) => (
                  <option key={conn.id} value={conn.id}>
                    {conn.github_account_name || conn.github_install_id || conn.id}
                  </option>
                ))}
              </select>
            </div>
          </>
        )}

        <div className="flex gap-2 justify-end mt-4">
          <Button
            type="button"
            onClick={handleClose}
            variant="secondary"
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={loading || vcsConnections.length === 0}
          >
            {loading ? 'Creating...' : 'Create Branch'}
          </Button>
        </div>
      </form>
    </ModalBase>
  )
}