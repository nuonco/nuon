'use client'

import React, { type FC, useState } from 'react'
import { FiMoreVertical } from 'react-icons/fi'
import { Link, Text } from '@/components'

export interface IDataTable {
  headers: Array<string>
  initData: Array<Array<React.ReactElement | string>>
}

export const DataTable: FC<IDataTable> = ({ headers, initData }) => {
  // const [data, setData] = useState(initData)
  const data = initData

  return (
    <div className="flex flex-col gap-8">
      <div className="flex-auto hidden">
        <input
          className="rounded bg-transparent border px-2 py-1"
          onInput={(e): void => {
            {
              /* const val = e.currentTarget.value
                const filteredData = initData.reduce(
                (acc: Array<Array<string>>, row) => {
                const match = row.some((item) => item?.includes(val))
                if (match) {
                acc.push(row)
                }
                return acc
                },
                []
                )

                setData(filteredData.length ? filteredData : initData) */
            }
          }}
          placeholder="Search"
          type="search"
        />
      </div>

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
                <td className="py-2" key={`cell-${i}`}>
                  {i + 1 !== row.length ? (
                    <>{td}</>
                  ) : (
                    <Link
                      className="text-gray-950 dark:text-gray-50"
                      href={td as string}
                    >
                      <FiMoreVertical />
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
