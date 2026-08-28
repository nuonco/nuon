import { useMutation } from '@tanstack/react-query'
import type { IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { createReleasePackageDownloadGrant } from '@/lib'
import type { TAPIError, TReleasePackage } from '@/types'
import { DownloadBundle } from './DownloadBundle'

export const DownloadBundleContainer = ({
  bundle: pkg,
  ...props
}: { bundle: TReleasePackage } & IButtonAsButton) => {
  const { org } = useOrg()
  const { addToast } = useToast()

  const { mutate: download, isPending } = useMutation({
    mutationFn: () =>
      createReleasePackageDownloadGrant({
        orgId: org!.id,
        packageId: pkg.id!,
      }),
    onSuccess: (grant) => {
      if (!grant?.url) {
        addToast(
          <Toast heading="Download failed" theme="error">
            <Text>The download grant did not include a URL.</Text>
          </Toast>
        )
        return
      }
      window.location.assign(grant.url)
      addToast(
        <Toast heading="Downloading bundle" theme="info">
          <Text>
            Downloading {grant.filename ?? pkg.id}. The download link expires
            shortly, so keep this tab open until it starts.
          </Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Download failed" theme="error">
          <Text>{err?.error || 'Unable to create a download grant.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DownloadBundle
      isPending={isPending}
      onClick={() => download()}
      {...props}
    />
  )
}
