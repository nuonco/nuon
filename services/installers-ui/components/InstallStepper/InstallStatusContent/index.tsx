import { Accordion, AccordionHeader, AccordionBody } from "../../Accordion";
import { InstallStatus } from "./InstallStatus";
import { StatusIcon } from "@/components";
import { InstallButton } from "./InstallButton";
import { Markdown } from "@/components/Markdown";

const interpolateInputsAndInstallId = (install_id, inputs: {}, text) => {
  let interpolated = text;
  Object.keys(inputs).forEach((key) => {
    let value = inputs[key];
    let regex = new RegExp(
      String.raw`\{\{[ ]{0,1}\.nuon\.install\.inputs\.${key}[ ]{0,1}\}\}`,
      "g",
    ); // https://regex101.com/r/U5iEei/1
    interpolated = interpolated.replaceAll(regex, value);
  });

  if (install_id) {
    // interpolate install.id
    let regex = new RegExp(
      String.raw`\{\{[ ]{0,1}\.nuon\.install\.id[ ]{0,1}\}\}`,
      "g",
    ); // https://regex101.com/r/U5iEei/1
    interpolated = interpolated.replaceAll(regex, install_id);
  }

  return interpolated;
};

export const InstallStatusContent = ({
  open = false,
  onClick = () => {},
  install = {
    id: null,
    status: "",
    status_description: "",
    install_inputs: [],
  },
  post_install_markdown = "",
}) => {
  let install_input_values = {};
  if (install.install_inputs && install.install_inputs.length > 0) {
    install_input_values =
      install.install_inputs[install.install_inputs.length - 1].values;
  }
  const post_install_markdown_with_inputs = interpolateInputsAndInstallId(
    install.id,
    install_input_values,
    post_install_markdown,
  );
  return (
    <Accordion open={open}>
      <AccordionHeader
        className={
          open
            ? `px-4
              text-accordion-header-active-color
              dark:text-accordion-header-active-color-dark
              hover:!text-gray-500
              bg-accordion-header-active-background
              dark:bg-accordion-header-active-background-dark`
            : "px-4"
        }
        onClick={onClick}
      >
        <span>
          Install Status <StatusIcon status={install.status} />{" "}
          <span className="text-sm font-medium">
            {install.status_description}
          </span>
        </span>
      </AccordionHeader>
      <AccordionBody>
        <div className="grid grid-cols-2 gap-2">
          <div>
            <Markdown content={post_install_markdown_with_inputs} />
          </div>

          <div>
            <InstallStatus install={install} />
            <div className="mt-4 flex justify-end">
              <InstallButton install={install} />
            </div>
          </div>
        </div>
      </AccordionBody>
    </Accordion>
  );
};
