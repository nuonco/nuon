// @ts-nocheck
import React, { type FC } from 'react'
import { ConfigurationVariables } from '@/components/old/ComponentConfig'
import type { TInstall, TInstallDeployPlanIntermediateData } from '@/types'

interface IInstallDeployIntermediateData {
  data: TInstallDeployPlanIntermediateData
  install?: TInstall
}

export const InstallDeployIntermediateData: FC<
  IInstallDeployIntermediateData
> = ({ data, install }) => {
  return (
    <>
      {data?.nuon?.components
      ? Object.keys(data?.nuon?.components).map((key) => (
        <ConfigurationVariables
          key={key}
          heading={`Component ${key}`}
          variables={data?.nuon?.components[key]?.outputs}
        />
      ))
      : null}
      {data?.nuon?.install?.internal_domain &&
       data?.nuon?.install?.public_domain ? (
         <ConfigurationVariables
           heading="Install domains"
           variables={{
             internal_domain: data?.nuon?.install?.internal_domain,
             public_domain: data?.nuon?.install?.public_domain,
           }}
         />
       ) : null}

      {install?.install_inputs?.[0]?.redacted_values ? (
        <ConfigurationVariables
          heading="Install inputs"
          variables={install?.install_inputs?.[0]?.redacted_values}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.account ? (
        <ConfigurationVariables
          heading="Sandbox outputs account"
          variables={data?.nuon?.install?.sandbox?.outputs?.account}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.cluster ? (
        <ConfigurationVariables
          heading="Sandbox outputs cluster"
          variables={data?.nuon?.install?.sandbox?.outputs?.cluster}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.ecr ? (
        <ConfigurationVariables
          heading="Sandbox outputs ECR"
          variables={data?.nuon?.install?.sandbox?.outputs?.ecr}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.internal_domain ? (
        <ConfigurationVariables
          heading="Sandbox outputs internal domain"
          variables={data?.nuon?.install?.sandbox?.outputs?.internal_domain}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.public_domain ? (
        <ConfigurationVariables
          heading="Sandbox outputs public domain"
          variables={data?.nuon?.install?.sandbox?.outputs?.public_domain}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.runner ? (
        <ConfigurationVariables
          heading="Sandbox outputs runner"
          variables={data?.nuon?.install?.sandbox?.outputs?.runner}
        />
      ) : null}

      {data?.nuon?.install?.sandbox?.outputs?.vpc ? (
        <ConfigurationVariables
          heading="Sandbox outputs VPC"
          variables={data?.nuon?.install?.sandbox?.outputs?.vpc}
        />
      ) : null}
    </>
  )
}
