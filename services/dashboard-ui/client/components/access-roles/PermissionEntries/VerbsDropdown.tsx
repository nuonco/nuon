import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import { Label } from '@/components/common/form/Label'
import type { TPermissionVerb } from '@/types'
import { ALL_VERBS, VERB_OPTIONS } from '../permissions'

export const VerbsDropdown = ({
  id,
  value,
  onChange,
  disabled,
}: {
  id: string
  value: TPermissionVerb[]
  onChange: (next: TPermissionVerb[]) => void
  disabled?: boolean
}) => {
  const toggle = (verb: TPermissionVerb, checked: boolean) =>
    onChange(
      checked
        ? ALL_VERBS.filter((v) => v === verb || value.includes(v))
        : value.filter((v) => v !== verb)
    )

  return (
    <div className="flex flex-col gap-1">
      <Label htmlFor={id}>
        <Text variant="body" className="font-medium">
          Allow
        </Text>
      </Label>
      <Dropdown
        id={id}
        alignment="left"
        closeOnBlur={false}
        disabled={disabled}
        variant="secondary"
        buttonClassName="w-full !justify-between"
        wrapperClassName="w-full"
        buttonText={summarize(value)}
      >
        <Menu className="min-w-56">
          {VERB_OPTIONS.map((verb) => (
            <div className="flex items-center space-x-2" key={verb.value}>
              <CheckboxInputWithButton
                buttonProps={{
                  className:
                    '!p-1 flex items-center justify-between group w-full',
                  children: (
                    <>
                      <span className="text-xs font-semibold">
                        {verb.label}
                      </span>
                      <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                        Only
                      </span>
                    </>
                  ),
                  type: 'button',
                  variant: 'ghost',
                  value: verb.value,
                  onClick: () => onChange([verb.value]),
                }}
                className="w-full"
                name={`${id}-${verb.value}`}
                value={verb.value}
                checked={value.includes(verb.value)}
                onChange={(e) => toggle(verb.value, e.target.checked)}
              />
            </div>
          ))}

          <hr />

          <Button
            className="w-full !p-1 shrink-0"
            type="button"
            onClick={() => onChange([...ALL_VERBS])}
            size="sm"
            variant="ghost"
          >
            Select all
          </Button>
        </Menu>
      </Dropdown>
    </div>
  )
}

function summarize(value: TPermissionVerb[]): string {
  if (value.length === 0) return 'Select actions'
  if (value.length === ALL_VERBS.length) return 'All actions'

  // Ordered by ALL_VERBS rather than selection order so the summary does not
  // reshuffle as boxes are ticked.
  return ALL_VERBS.filter((verb) => value.includes(verb))
    .map((verb) => verb.charAt(0).toUpperCase() + verb.slice(1))
    .join(', ')
}
