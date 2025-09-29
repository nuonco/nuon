import { Notice, Text, Markdown } from '@/components'
import { getInstallReadme } from '@/lib'

export const Readme = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const { data: installReadme, error } = await getInstallReadme({
    installId,
    orgId,
  })

  return installReadme && !error ? (
    <div className="flex flex-col gap-3">
      {installReadme?.warnings?.length
        ? installReadme?.warnings?.map((warn, i) => (
            <Notice key={i.toString()} variant="warn">
              {warn}
            </Notice>
          ))
        : null}
      <Markdown content={installReadme?.readme} />
    </div>
  ) : (
    <Text variant="reg-12">No install README found</Text>
  )
}
