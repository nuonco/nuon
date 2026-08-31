import { useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { Block } from './Block'
import { Drawer } from './Drawer'
import { appById } from './fixtures'
import { rowHoverClass } from './utils'

export const BranchSwitcher = () => {
  const { appId = '', branchId = '' } = useParams()
  const navigate = useNavigate()
  const [isOpen, setIsOpen] = useState(false)

  const app = appById(appId)
  const current = app.branches.find((branch) => branch.id === branchId)

  return (
    <>
      <Block
        className="h-[32px] cursor-pointer"
        title="Switch branch"
        icon="GitBranchIcon"
        text={current?.label ?? 'branch'}
        onClick={() => setIsOpen(true)}
      />

      {isOpen && (
        <Drawer title="Branches" onClose={() => setIsOpen(false)}>
          <div className="flex flex-col gap-2">
            {app.branches.map((branch) => (
              <div
                key={branch.id}
                className={`flex items-center justify-between gap-4 ${rowHoverClass}`}
                title={branch.label}
                onClick={() => {
                  setIsOpen(false)
                  navigate(`/apps/${appId}/branches/${branch.id}`)
                }}
              >
                <Block
                  className="h-[12px] cursor-pointer"
                  icon="GitBranchIcon"
                  iconSize={12}
                  text={branch.label}
                />
                {branch.id === branchId && (
                  <Block
                    className="h-[10px] flex-none"
                    icon="CheckIcon"
                    iconSize={12}
                  />
                )}
              </div>
            ))}
          </div>

          <Block
            className="h-[32px] w-full cursor-pointer opacity-60"
            title="Manage branches"
            icon="SlidersHorizontalIcon"
            text="Manage branches"
          />
        </Drawer>
      )}
    </>
  )
}
