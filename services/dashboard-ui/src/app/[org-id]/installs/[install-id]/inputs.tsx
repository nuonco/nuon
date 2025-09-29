import {
  InstallInputs,
  InstallInputsModal,
  SectionHeader,
  Text,
} from '@/components'
import { getInstallCurrentInputs } from '@/lib'

export const CurrentInputs = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const { data: currentInputs } = await getInstallCurrentInputs({
    installId,
    orgId,
  })

  return (
    <>
      <SectionHeader
        actions={
          currentInputs?.redacted_values ? (
            <InstallInputsModal currentInputs={currentInputs} />
          ) : undefined
        }
        className="mb-4"
        heading="Current inputs"
      />
      {currentInputs?.redacted_values ? (
        <InstallInputs currentInputs={currentInputs} />
      ) : (
        <Text>No inputs configured.</Text>
      )}
    </>
  )
}
