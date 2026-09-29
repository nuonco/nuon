import { z } from 'zod'
import { APP_NAME_PATTERN, APP_NAME_RULE, isCloud, isKnownRegion, type TCloud } from './constants'

export const startSchema = z.object({
  appName: z
    .string()
    .min(1, 'Name your app template to continue.')
    .regex(APP_NAME_PATTERN, APP_NAME_RULE),
  cloud: z.string().refine((value) => isCloud(value), 'Select a test cloud to continue.'),
})

export type StartValues = z.input<typeof startSchema>

export const deploySchema = (cloud: TCloud) =>
  z.object({
    region: z.string().refine((region) => isKnownRegion(cloud, region), 'Choose a region'),
    autoApprove: z.boolean(),
  })

export type DeployValues = z.infer<ReturnType<typeof deploySchema>>
