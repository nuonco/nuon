import { useState } from 'react'
import { useParams } from 'react-router'
import { Block } from './Block'
import { Drawer } from './Drawer'
import { LinkBlock } from './LinkBlock'
import { runnables } from './fixtures'
import { labelWidth, rowHoverClass } from './utils'

const pinned = runnables.filter((runnable) => runnable.pinned)

const groups = [
  { title: 'Actions', kind: 'action' as const },
  { title: 'Runbooks', kind: 'runbook' as const },
]

export const RunPicker = () => {
  const { installId = '' } = useParams()
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="flex flex-wrap items-center gap-3">
      {pinned.map((runnable) => (
        <Block
          key={runnable.label}
          className="h-[32px] cursor-pointer"
          style={{ width: labelWidth(runnable.label) }}
          title={runnable.label}
          icon="PlayIcon"
          text={runnable.label}
        />
      ))}

      <Block
        className="h-[32px] cursor-pointer opacity-60"
        style={{ width: 140 }}
        title="Browse all"
        icon="DotsThreeIcon"
        text={`All ${runnables.length}`}
        onClick={() => setIsOpen(true)}
      />

      {isOpen && (
        <Drawer title="Run" onClose={() => setIsOpen(false)}>
          <Block
            className="h-[32px] w-full"
            title="Search"
            icon="MagnifyingGlassIcon"
            text="Search actions and runbooks"
          />

          {groups.map((group) => (
            <div key={group.title} className="flex flex-col gap-2">
              <Block
                className="h-[8px] opacity-50"
                style={{ width: labelWidth(group.title) }}
                title={group.title}
                text={group.title}
              />

              {runnables
                .filter((runnable) => runnable.kind === group.kind)
                .map((runnable) => (
                  <div
                    key={runnable.label}
                    className={`flex items-center justify-between gap-4 ${rowHoverClass}`}
                    title={runnable.label}
                  >
                    <Block
                      className="h-[12px]"
                      icon="PlayIcon"
                      iconSize={12}
                      text={runnable.label}
                    />
                    <LinkBlock
                      path={`/installs/${installId}/${runnable.kind === 'runbook' ? 'runbooks' : 'actions'}/${runnable.label.toLowerCase().replace(/ /g, '-')}`}
                      label="Open"
                      className="h-[10px] flex-none opacity-50"
                    />
                  </div>
                ))}
            </div>
          ))}
        </Drawer>
      )}
    </div>
  )
}
