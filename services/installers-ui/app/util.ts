export function installRequestBody(
  app: Record<string, any>,
  formData: FormData,
) {
  const data = Object.fromEntries(formData);
  const inputs = Object.keys(data).reduce(
    (acc: Record<string, unknown>, key) => {
      if (key.includes("input:")) {
        acc[key.replace("input:", "")] = data[key];
      }

      return acc;
    },
    {},
  );

  let input: Record<string, any> = {
    inputs,
    name: data?.name,
  };

  if (app?.cloud_platform === "azure") {
    input = {
      azure_account: {
        location: data?.location,
        service_principal_app_id: data?.service_principal_app_id,
        service_principal_password: data?.service_principal_password,
        subscription_id: data?.subscription_id,
        subscription_tenant_id: data?.subscription_tenant_id,
      },
      ...input,
    };
  }

  if (app?.cloud_platform === "aws") {
    input = {
      aws_account: {
        iam_role_arn: data?.iam_role_arn,
        region: data?.region,
      },
      ...input,
    };
  }

  return input;
}
