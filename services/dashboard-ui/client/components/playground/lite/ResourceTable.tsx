import { Block } from './Block'
import { LinkBlock } from './LinkBlock'
import { Toolbar } from './Toolbar'
import { rowHoverClass } from './utils'

export interface IResourceRow {
  id: string
  label: string
  widths: number[]
}

export interface IResourceTable {
  headers: string[]
  rows: IResourceRow[]
  filters?: string[]
  basePath?: string
  statusColumn?: number
}

export const makeRows = (
  prefix: string,
  count: number,
  label: string
): IResourceRow[] =>
  Array.from({ length: count }, (_, i) => ({
    id: `${prefix}-${String(i + 1).padStart(2, '0')}`,
    label: `${label} ${i + 1}`,
    widths: [110 + ((i * 37) % 90), 80 + ((i * 23) % 60), 70 + ((i * 17) % 40)],
  }))

export const ResourceTable = ({
  headers,
  rows,
  filters = ['Status'],
  basePath,
  statusColumn = 1,
}: IResourceTable) => {
  const gridClass = `grid gap-5 items-center`
  const template = `minmax(0,2fr) ${headers
    .slice(1)
    .map(() => 'minmax(0,1fr)')
    .join(' ')}`

  return (
    <div className="flex flex-col gap-4">
      <Toolbar filters={filters} />

      <div className="flex flex-col gap-3">
        <div
          className={`${gridClass} pb-1`}
          style={{ gridTemplateColumns: template }}
        >
          {headers.map((header) => (
            <Block
              key={header}
              className="h-[8px] w-[56px] max-w-full opacity-50"
              title={header}
              text={header}
            />
          ))}
        </div>

        {rows.map((row) => (
          <div
            key={row.id}
            className={`${gridClass} ${rowHoverClass}`}
            style={{ gridTemplateColumns: template }}
          >
            <div className="flex min-w-0 flex-col gap-1.5">
              {basePath ? (
                <LinkBlock
                  path={`${basePath}/${row.id}`}
                  label={row.label}
                  className="h-[12px] max-w-full"
                  style={{ width: row.widths[0] }}
                />
              ) : (
                <Block
                  className="h-[12px] max-w-full"
                  text={row.label}
                  style={{ width: row.widths[0] }}
                />
              )}
              <Block
                className="h-[8px] w-[96px] max-w-full opacity-50"
                title={row.id}
                text={row.id}
              />
            </div>

            {headers
              .slice(1)
              .map((header, i) =>
                i === statusColumn - 1 ? (
                  <Block
                    key={header}
                    className="h-[16px] w-[64px] max-w-full rounded-full"
                  />
                ) : (
                  <Block
                    key={header}
                    className="h-[10px] max-w-full opacity-60"
                    style={{ width: row.widths[(i % 2) + 1] }}
                  />
                )
              )}
          </div>
        ))}
      </div>
    </div>
  )
}
