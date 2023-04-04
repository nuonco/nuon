import { readFileSync } from "node:fs";
import { getBooleanInput, getInput, info, setFailed } from "@actions/core";
import { rmRF } from "@actions/io";
import ensureError from "ensure-error";
import { parse } from "./parse_output";

const run = async () => {
  try {
    const inFile = getInput("file", { required: true });
    const cleanup = getBooleanInput("cleanup");

    info(`Parsing terraform outputs from ${inFile}...`);
    const tfoutput = JSON.parse(readFileSync(inFile, "utf8"));
    parse(tfoutput);

    if (cleanup) {
      info(`cleaning up ${inFile}...`);
      rmRF(inFile);
    }
  } catch (_error: unknown) {
    const error = ensureError(_error);
    setFailed(error);
  }
};

void run();
