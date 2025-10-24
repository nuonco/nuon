'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'

export interface IDataTable {
  headers: Array<string>
  initData: Array<Array<React.ReactElement | string>>
}

export const DataTable: FC<IDataTable> = ({ headers, initData }) => {
  const data = initData

  return (
    <div className="overflow-auto">
      <div
        className="grid"
        style={{ gridTemplateColumns: `repeat(${headers.length}, auto)` }}
      >
        {headers.map((header, i) => (
          <div className={`py-4 ${i !== 0 && 'pl-6'}`} key={`header-${i}`}>
            <Text isMuted>{header}</Text>
          </div>
        ))}

        {data.map((row) =>
          row.map((td, i) => (
            <div
              className={`border-t py-4 ${i !== 0 && 'pl-6'}`}
              key={`cell-${i}`}
            >
              {td}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
