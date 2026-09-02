import { useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { Drawer } from './Drawer'
import { configRuns } from './fixtures'
import { branchBase } from './nav'
import { rowHoverClass } from './utils'

export interface ICompareDrawer {
  onClose: () => void
}

export const CompareDrawer = ({ onClose }: ICompareDrawer) => {
  const { appId = '', branchId = '' } = useParams()
  const navigate = useNavigate()
  const [base, setBase] = useState<string | undefined>(configRuns[0]?.id)
  const [compare, setCompare] = useState<string | undefined>()

  const select = (id: string) => {
    if (id === base) return setBase(undefined)
    if (id === compare) return setCompare(undefined)
    if (!base) return setBase(id)
    return setCompare(id)
  }

  const canCompare = Boolean(base && compare)

  return (
    <Drawer title="Compare config" onClose={onClose}>
      <div className="flex items-center gap-3">
        <Block
          className="h-[32px] flex-1"
          title="Base"
          text={base ?? 'Select base'}
        />
        <Block
          className="h-[12px] flex-none opacity-40"
          icon="ArrowRightIcon"
          iconSize={14}
        />
        <Block
          className="h-[32px] flex-1"
          title="Compare"
          text={compare ?? 'Select run'}
        />
      </div>

      <div className="flex flex-col gap-2">
        {configRuns.map((run) => {
          const isBase = run.id === base
          const isCompare = run.id === compare

          return (
            <div
              key={run.id}
              title={run.id}
              onClick={() => select(run.id)}
              className={cn(
                'flex items-center gap-3',
                rowHoverClass,
                (isBase || isCompare) &&
                  'bg-cool-grey-200 dark:bg-dark-grey-700'
              )}
            >
              <Block
                className="h-[12px] w-[12px] flex-none rounded-full"
                icon={isBase || isCompare ? 'CheckIcon' : undefined}
                iconSize={12}
              />
              <div className="flex min-w-0 flex-1 flex-col gap-1.5">
                <Block
                  className="h-[10px] cursor-pointer"
                  text={run.id}
                  style={{ width: run.labelWidth }}
                />
                <Block className="h-[8px] w-[110px] opacity-50" />
              </div>
              {(isBase || isCompare) && (
                <Block
                  className="h-[8px] flex-none opacity-60"
                  text={isBase ? 'base' : 'compare'}
                />
              )}
            </div>
          )
        })}
      </div>

      <Block
        className={cn(
          'h-[32px] w-full',
          canCompare ? 'cursor-pointer' : 'opacity-40'
        )}
        title={canCompare ? 'Compare' : 'Select two runs'}
        icon="SplitHorizontalIcon"
        text="Compare"
        onClick={() => {
          if (!canCompare) return
          onClose()
          navigate(`${branchBase(appId, branchId)}/compare/${base}/${compare}`)
        }}
      />
    </Drawer>
  )
}
