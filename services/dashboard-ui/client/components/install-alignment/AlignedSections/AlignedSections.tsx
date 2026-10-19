import type { ReactNode } from 'react'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TInstallAlignmentCategory } from '@/types'

type TAlignedItem = { id: string; alignment?: TInstallAlignmentCategory }

export interface IAlignedSections<T extends TAlignedItem> {
  items: T[]
  noun: string
  render: (items: T[]) => ReactNode
}

const isDriftCategory = (category?: TInstallAlignmentCategory) =>
  category === 'changes-not-applied' || category === 'state-drift-detected'

export const AlignedSections = <T extends TAlignedItem>({
  items,
  noun,
  render,
}: IAlignedSections<T>) => {
  const inSync = items.filter((item) => item.alignment === 'in-sync')
  const disabled = items.filter((item) => item.alignment === 'disabled')
  const other = items.filter(
    (item) =>
      item.alignment !== 'in-sync' &&
      item.alignment !== 'disabled' &&
      !isDriftCategory(item.alignment)
  )

  return (
    <>
      {inSync.length ? (
        <section
          className="flex flex-col gap-4"
          aria-label={noun === 'images' ? 'Synced' : 'In sync'}
          data-testid="in-sync-section"
        >
          <SectionHeader
            title={noun === 'images' ? 'Synced' : 'In sync'}
            description={
              noun === 'images'
                ? 'The configured image versions have been synced to this install’s registry. This does not verify which versions workloads are running.'
                : `These ${noun} match the desired configuration and inputs. No changes are needed.`
            }
          />
          {render(inSync)}
        </section>
      ) : null}
      {other.length ? (
        noun === 'images' ? (
          render(other)
        ) : (
          <section
            className="flex flex-col gap-4"
            aria-label="Not compared"
            data-testid="not-compared-section"
          >
            <SectionHeader
              title="Not compared"
              description={`Drift could not be determined for these ${noun}.`}
            />
            {render(other)}
          </section>
        )
      ) : null}
      {disabled.length ? (
        <section
          className="flex flex-col gap-3"
          aria-label={`Disabled ${noun}`}
          data-testid="disabled-section"
        >
          <SectionHeader title={`Disabled ${noun}`} />
          {render(disabled)}
        </section>
      ) : null}
    </>
  )
}
