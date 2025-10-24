// import type { TAppSandboxConfig } from '@/types'

// export function composeCloudFormationQuickCreateUrl(sandboxConfig: TAppSandboxConfig): string {
//   const aws_delegation_config = sandboxConfig?.aws_delegation_config
//   const base_url =
//     'https://us-west-2.console.aws.amazon.com/cloudformation/home#/stacks/quickcreate'
//   const delegationEnabled =
//     Boolean(aws_delegation_config && aws_delegation_config?.iam_role_arn)
//   let params = {}

//   if (delegationEnabled) {
//     params = {
//       templateUrl:
//         'https://nuon-artifacts.s3.us-west-2.amazonaws.com/sandbox/aws-ecs/cloudformation-template-delegation.yaml',
//       stackName: `nuon-${sandboxConfig?.public_git_vcs_config?.directory}-permissions`,
//       param_DelegationRoleARN: `${aws_delegation_config?.iam_role_arn}`,
//     }
//   } else {
//     params = {
//       templateUrl:
//         sandboxConfig?.artifacts?.cloudformation_stack_template,
//       stackName: `nuon-${sandboxConfig?.public_git_vcs_config?.directory}-permissions`,    
//     }
//   }

//   let searchParams = new URLSearchParams(params)
//   let url = new URL(base_url)
//   return url + '?' + searchParams.toString()
// }
