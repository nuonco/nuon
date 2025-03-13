'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { BookOpen } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CodeViewer } from '@/components/Code'
import { Table } from '@/components/DataTable'
import { CheckboxInput } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import type { TInstall } from '@/types'

type TPermissionOpetion = {
  name: string
  display_name: string
  descriptions: string
  perms: Record<string, string>
}

const PERMISSION_OPTIONS: Array<TPermissionOpetion> = [
  {
    display_name: 'Account access',
    descriptions: 'Permissions to interact with cloud provisioning.',
    name: 'account-access',
    perms: { something: 'permissions dude' },
  },
  {
    display_name: 'Sandbox access',
    descriptions: 'Permission to interact with Sandbox enviornment',
    name: 'sandbox-access',
    perms: { some: 'thing' },
  },
  {
    display_name: 'Runner access',
    descriptions: 'Permission to interact with runner.',
    name: 'runner-access',
    perms: { thing: 'stuff' },
  },
  {
    display_name: 'Deprovision install access',
    descriptions: 'Permissions to deprocision the existing install.',
    name: 'deprovision-access',
    perms: { thing: 'perms ' },
  },
]

interface IBreakGlassForm {
  install: TInstall
}

export const BreakGlassForm: FC<IBreakGlassForm> = ({ install }) => {
  const [columnFilters, _] = useState([])
  const [globalFilter, __] = useState('')
  const columns: Array<ColumnDef<TPermissionOpetion>> = useMemo(
    () => [
      {
        header: 'List of permissions',
        accessorKey: 'name',
        cell: (props) => (
          <CheckboxInput
            className="w-fit"
            name={props.getValue<string>()}
            labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-8 w-fit"
            labelText={
              <span className="flex flex-col gap2">
                <Text variant="med-14">
                  {props.row?.original?.display_name}
                </Text>
                <Text className="text-cool-grey-600 dark:text-white/70">
                  {props.row.original.descriptions}
                </Text>
              </span>
            }
          />
        ),
      },

      {
        header: 'Controls',
        accessorKey: 'perms',
        cell: (props) => <PermissionModal {...props.row.original} />,
      },
    ],
    []
  )

  return (
    <div>
      <form>
        <Table
          columns={columns}
          data={PERMISSION_OPTIONS}
          columnFilters={columnFilters}
          emptyMessage="Reset your search and try again."
          emptyTitle="No components found"
          globalFilter={globalFilter}
        />
        <div className="border-t py-4">
          <div className="flex items-start justify-between">
            <CheckboxInput
              name="ack"
              labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-8 max-w-[650px]"
              labelText={
                <span className="flex flex-col gap2">
                  <Text variant="med-14">
                    I acknowledge that I understand and accept the changes to
                    the installation permissions.
                  </Text>
                  <Text className="!font-normal" variant="reg-12">
                    I recognize that these modifications may impact access,
                    permissions, and resource management, and I agree to use
                    these settings responsibly and in compliance with
                    organizational policies.
                  </Text>
                </span>
              }
            />

            <Button className="text-sm !font-medium" variant="primary">
              Generate CloudFormation stack
            </Button>
          </div>
        </div>
      </form>
    </div>
  )
}

const PermissionModal: FC<TPermissionOpetion> = ({
  display_name,
  descriptions,
  perms,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <>
      <Modal
        className="!max-w-4xl"
        heading={
          <span className="flex flex-col gap-1">
            <Text variant="med-14">{display_name}</Text>
            <Text className="!font-normal">{descriptions}</Text>
          </span>
        }
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false)
        }}
      >
        <CodeViewer
          initCodeSource={JSON.stringify(perms, null, 2)}
          language="json"
        />
      </Modal>
      <Button
        className="text-sm !font-medium flex items-center gap-3"
        onClick={() => {
          setIsOpen(true)
        }}
        type="button"
      >
        <BookOpen /> View permissions
      </Button>
    </>
  )
}
