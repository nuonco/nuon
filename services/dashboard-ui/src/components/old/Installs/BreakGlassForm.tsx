'use client'

import React, { type FC, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { type ColumnDef } from '@tanstack/react-table'
import { BookOpen } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { ClickToCopyButton } from '@/components/old/ClickToCopy'
import { CodeViewer } from '@/components/old/Code'
import { Table } from '@/components/old/DataTable'
import { CheckboxInput } from '@/components/old/Input'
import { Modal } from '@/components/old/Modal'
import { ToolTip } from '@/components/old/ToolTip'
import { Code, Text } from '@/components/old/Typography'
import type { TInstall } from '@/types'

type TPermissionOpetion = {
  name: string
  display_name: string
  descriptions: string
  perms: Record<string, unknown>
}

const PERMISSION_OPTIONS: Array<TPermissionOpetion> = [
  {
    display_name: 'Network access',
    descriptions:
      'Permissions to access, update, and delete the VPC the install is provisioned in.',
    name: 'network-access',
    perms: {
      Version: '2012-10-17',
      Statement: [
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::676549690856:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::007754799877:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::814326426574:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::766121324316:root',
          },
          Action: ['sts:AssumeRole'],
        },
      ],
    },
  },
  {
    display_name: 'Sandbox access',
    descriptions:
      'Permissions to access, update, and delete the install sandbox resources.',
    name: 'sandbox-access',
    perms: {
      Version: '2012-10-17',
      Statement: [
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::676549690856:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::007754799877:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::814326426574:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::766121324316:root',
          },
          Action: ['sts:AssumeRole'],
        },
      ],
    },
  },
  {
    display_name: 'Runner access',
    descriptions:
      'Permissions to access, update, and delete the install runner resources.',
    name: 'runner-access',
    perms: {
      Version: '2012-10-17',
      Statement: [
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::676549690856:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::007754799877:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::814326426574:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::766121324316:root',
          },
          Action: ['sts:AssumeRole'],
        },
      ],
    },
  },
  {
    display_name: 'Component access',
    descriptions:
      'Permissions to access, update, and delete the install component resources.',
    name: 'component-access',
    perms: {
      Version: '2012-10-17',
      Statement: [
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::676549690856:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::007754799877:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::814326426574:root',
          },
          Action: ['sts:AssumeRole'],
        },
        {
          Sid: '',
          Effect: 'Allow',
          Principal: {
            AWS: 'arn:aws:iam::766121324316:root',
          },
          Action: ['sts:AssumeRole'],
        },
      ],
    },
  },
]

interface IBreakGlassForm {
  install: TInstall
}

export const BreakGlassForm: FC<IBreakGlassForm> = ({ install }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [cfLink, setCFLink] = useState(
    'https://us-east-1.console.aws.amazon.com/cloudformation/home?region=us-east-1#'
  )
  const [columnFilters, _] = useState([])
  const [globalFilter, __] = useState('')
  const [accept, setAccept] = useState(false)
  const [accessLevels, setAccessLevels] = useState({
    'network-access': false,
    'sandbox-access': false,
    'runner-access': false,
    'component-access': false,
  })
  const columns: Array<ColumnDef<TPermissionOpetion>> = useMemo(
    () => [
      {
        header: 'List of permissions',
        accessorKey: 'name',
        cell: (props) => (
          <CheckboxInput
            className="w-fit"
            defaultChecked={accessLevels[props.row.original.name]}
            name={props.getValue<string>()}
            labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-8 w-fit"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              setAccessLevels((state) => ({
                ...state,
                [props.row.original.name]: Boolean(e?.target?.checked),
              }))
            }}
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
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-2xl"
              contentClassName="!p-0"
              heading="CloudFormation stack created"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="p-6">
                <div className="border rounded-lg p-3 flex flex-col gap-3">
                  <div className="flex justify-between items-center">
                    <Text variant="med-14">
                      <ToolTip tipContent="Not sure what needs to go here">
                        CloudFormation Stack
                      </ToolTip>
                    </Text>
                    <ClickToCopyButton
                      className="border rounded-md p-1 text-sm"
                      textToCopy={cfLink}
                    />
                  </div>
                  <Code>{cfLink}</Code>
                </div>
              </div>
              <div className="p-6 border-t flex justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                  variant="primary"
                >
                  Done
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <form
        onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
          e.preventDefault()

          // TODO(nnnat): create cloudformation link

          setIsOpen(true)
        }}
      >
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
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setAccept(Boolean(e?.currentTarget?.checked))
              }}
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

            <Button
              disabled={
                !Boolean(
                  accept &&
                    Object.values(accessLevels).some((lvl) => lvl === true)
                )
              }
              className="text-sm !font-medium"
              variant="primary"
            >
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
      {isOpen
        ? createPortal(
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
            </Modal>,
            document.body
          )
        : null}
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
