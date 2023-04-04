import { debug, exportVariable, setSecret } from "@actions/core";

type terraformOutput = {
  sensitive: boolean;
  type: string;
  value: string | any; // eslint-disable-line @typescript-eslint/no-explicit-any
};
type terraformOutputs = Record<string, terraformOutput>;

export function parse(outputs: terraformOutputs): void {
  Object.entries(outputs).forEach(([k, v]) => {
    if (v.type === "string") {
      debug(`got top level string for key: ${k}`);
      setEnv(v.sensitive, k, v.value);
      return;
    }

    debug(`got top level object for key: ${k}`);
    parseNested(v.sensitive, k, v.value);
  });
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const parseNested = (sensitive: boolean, prefix: string, nested: any): void => {
  Object.entries(nested).forEach(([k, v]) => {
    const name = `${prefix}_${k}`;
    if (typeof v === "string") {
      debug(`got nested string for key: ${name}`);
      setEnv(sensitive, name, v);
      return;
    }

    if (Array.isArray(v)) {
      debug(`got nested array for key: ${name}`);
      v.forEach((v, i) => {
        parseNested(sensitive, `${name}_${i}`, v);
      });
      return;
    }

    debug(`got nested object for key: ${name}`);
    parseNested(sensitive, name, v);
  });
};

const setEnv = (sensitive: boolean, key: string, value: string): void => {
  if (sensitive) {
    setSecret(value);
  }
  const k = `TFO_${key.toUpperCase().replaceAll("-", "_")}`;
  exportVariable(k, value);
  debug(`set: ${k} = ${value}`);
};
