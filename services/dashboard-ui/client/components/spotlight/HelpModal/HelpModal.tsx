import { useMemo } from 'react'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { toSentenceCase } from '@/utils/string-utils'
import { COMMANDS_BY_PREFIX, COMMAND_DESCRIPTIONS, FILTER_PREFIXES } from '../types'

const Kbd = ({ children }: { children: React.ReactNode }) => (
  <kbd className="inline-flex items-center justify-center min-w-[20px] h-5 px-1.5 border border-[rgba(158,168,179,0.24)] rounded bg-white dark:bg-dark-grey-900 text-[11px] font-medium text-[#9ea8b3]">
    {children}
  </kbd>
)

const Keys = ({ children }: { children: React.ReactNode }) => (
  <span className="flex items-center gap-1">{children}</span>
)

const isMac = () =>
  typeof navigator !== 'undefined' &&
  /Mac|iPhone|iPad|iPod/.test(navigator.userAgent)

export const HelpModal = ({ ...props }: IModal) => {
  const mac = useMemo(isMac, [])
  const mod = mac ? '⌘' : 'Ctrl'
  const alt = mac ? '⌥' : 'Alt'
  const shift = mac ? '⇧' : 'Shift'

  const shortcuts = useMemo(
    () => [
      {
        shortcut: 'Open spotlight search',
        keys: (
          <Keys>
            <Kbd>{mod}</Kbd>
            <Kbd>K</Kbd>
          </Keys>
        ),
      },
      {
        shortcut: 'Open help',
        keys: (
          <Keys>
            <Kbd>{mod}</Kbd>
            <Kbd>H</Kbd>
          </Keys>
        ),
      },
      {
        shortcut: 'Toggle sidebar',
        keys: (
          <Keys>
            <Kbd>{alt}</Kbd>
            <Kbd>S</Kbd>
          </Keys>
        ),
      },
      {
        shortcut: 'Toggle page sidebar',
        keys: (
          <Keys>
            <Kbd>{alt}</Kbd>
            <Kbd>{shift}</Kbd>
            <Kbd>S</Kbd>
          </Keys>
        ),
      },
      {
        shortcut: 'Close modal or panel',
        keys: (
          <Keys>
            <Kbd>Esc</Kbd>
          </Keys>
        ),
      },
      {
        shortcut: 'Cycle log lines (when log panel is open)',
        keys: (
          <Keys>
            <Kbd>↑</Kbd> <Kbd>↓</Kbd> or <Kbd>K</Kbd> <Kbd>J</Kbd>
          </Keys>
        ),
      },
    ],
    [mod, alt, shift]
  )

  const commandEntries = Object.entries(COMMANDS_BY_PREFIX) as [
    string,
    string[],
  ][]

  return (
    <Modal
      heading="Keyboard shortcuts & commands"
      size="lg"
      showFooter={false}
      {...props}
    >
      <div className="flex flex-col gap-8">
        <section>
          <div className="mb-3">
            <Text variant="base" weight="strong">
              Keyboard shortcuts
            </Text>
          </div>
          <PropertyGrid
            values={shortcuts}
            columns={[
              { key: 'shortcut', header: 'Action' },
              {
                key: 'keys',
                header: 'Shortcut',
                className: 'text-right justify-end',
              },
            ]}
          />
        </section>

        <section>
          <div className="mb-2">
            <Text variant="base" weight="strong">
              Spotlight search
            </Text>
          </div>
          <div className="mb-3">
            <Text variant="subtext" theme="neutral">
              Press <Kbd>{mod}</Kbd> <Kbd>K</Kbd> to open, then type to search.
              Use a prefix to scope results, press <Kbd>Tab</Kbd> to
              autocomplete.
            </Text>
          </div>
          <div className="flex items-center gap-1.5 flex-wrap">
            {FILTER_PREFIXES.map((prefix) => (
              <Badge key={prefix} size="sm" variant="code" theme="neutral">
                {prefix}
              </Badge>
            ))}
          </div>
        </section>

        <section>
          <div className="mb-2">
            <Text variant="base" weight="strong">
              Slash commands
            </Text>
          </div>
          <div className="mb-4">
            <Text variant="subtext" theme="neutral">
              Type <Kbd>/</Kbd> after a filter to run commands on the matched
              resource, e.g.{' '}
            </Text>
            <Badge className="mt-1" size="sm" variant="code" theme="neutral">
              install:my-install/deploy all components
            </Badge>
          </div>
          <div className="flex flex-col gap-6 ">
            {commandEntries.map(([prefix, commands]) => (
              <div key={prefix}>
                <div className="flex items-center gap-4 mb-2">
                  <Text>{toSentenceCase(prefix)}</Text>
                  <Badge size="sm" variant="code" theme="brand">
                    {prefix}:
                  </Badge>
                </div>
                <PropertyGrid
                  values={commands.map((cmd) => ({
                    command: (
                      <Text variant="subtext" family="mono">{`/${cmd}`}</Text>
                    ),
                    description: COMMAND_DESCRIPTIONS[cmd] ?? '',
                  }))}
                  columns={[
                    { key: 'command', header: 'Command' },
                    { key: 'description', header: 'Description' },
                  ]}
                />
              </div>
            ))}
          </div>
        </section>
      </div>
    </Modal>
  )
}
