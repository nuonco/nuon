'use client'

import React, { type FC } from 'react'
import { Link } from '@/components/Link'
import { CaretRight } from '@phosphor-icons/react'

export interface IDataTable {
  headers: Array<string>
  initData: Array<Array<React.ReactElement | string>>
}

export const DataTable: FC<IDataTable> = ({ headers, initData }) => {
  const data = initData

  return (
    <div className="flex flex-col gap-8">
      <table className="table-auto w-full">
        <thead>
          <tr className="border-b text-left">
            {headers.map((header, i) => (
              <th className="text-sm" key={`header-${i}`}>
                {header}
              </th>
            ))}
            <th></th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {data.map((row, i) => (
            <tr key={`row-${i}`}>
              {row.map((td, i) => (
                <td className="py-2 max-w-24 overflow-hidden" key={`cell-${i}`}>
                  {i + 1 !== row.length ? (
                    <>{td}</>
                  ) : (
                    <Link
                      className="text-gray-950 dark:text-gray-50"
                      href={td as string}
                    >
                      <CaretRight />
                    </Link>
                  )}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
