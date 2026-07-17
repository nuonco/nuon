// Express router that proxies the customer-facing install surface to ctl-api.
// Mount with: app.use("/api", installsRouter)
//
// Security invariants (see references/architecture-and-security.md):
//   - Nuon token/org stay server-side (in nuonClient).
//   - Every route authorizes the customer before forwarding.
//   - Customer input is whitelisted to the app's declared inputs.
//   - app_id + cloud account + tenant labels are server-owned.

const express = require("express");
const { nuon, NuonError } = require("./nuonClient");

const APP_ID = process.env.NUON_APP_ID;
const CLOUD_PLATFORM = process.env.NUON_CLOUD_PLATFORM || "aws"; // aws|azure|gcp

const router = express.Router();

// TODO: replace with your real session/tenant model.
function getCustomer(req) {
  // e.g. return req.session.customer
  return req.customer;
}

// TODO: enforce your tenant → app authorization.
function authorizeCreate(customer /*, appId */) {
  if (!customer) {
    const e = new Error("unauthorized");
    e.status = 401;
    throw e;
  }
  // e.g. assert customer may create installs for APP_ID
}

// TODO: enforce that this customer owns this install.
function authorizeReadInstall(customer /*, installId */) {
  if (!customer) {
    const e = new Error("unauthorized");
    e.status = 401;
    throw e;
  }
}

// Only inputs the vendor marked `source: "customer"` are shown to and accepted
// from the customer. `vendor` inputs (the default) resolve from the app config
// server-side and must never be settable by the browser.
function customerFacingInputs(schema) {
  return (schema.inputs || []).filter((i) => i.source === "customer" && !i.internal);
}

function mapNuonError(err, res) {
  if (err instanceof NuonError) {
    console.error("ctl-api error", err.status, err.body);
    const map = { 400: 422, 401: 502, 403: 502, 404: 404, 409: 409 };
    const status = map[err.status] || 502;
    return res.status(status).json({ error: "install request failed" });
  }
  const status = err.status || 500;
  return res.status(status).json({ error: err.message || "server error" });
}

// GET /api/install-inputs — dynamic form schema (internal inputs stripped)
router.get("/install-inputs", async (req, res) => {
  try {
    authorizeReadInstall(getCustomer(req));
    const schema = await nuon.getAppInputSchema(APP_ID);
    res.json({ inputs: customerFacingInputs(schema), input_groups: schema.input_groups || [] });
  } catch (err) {
    mapNuonError(err, res);
  }
});

// POST /api/installs — create an install
router.post("/installs", async (req, res) => {
  try {
    const customer = getCustomer(req);
    authorizeCreate(customer, APP_ID);

    const { name, inputs = {} } = req.body || {};
    if (typeof name !== "string" || !name.trim()) {
      return res.status(422).json({ error: "name is required" });
    }

    const schema = await nuon.getAppInputSchema(APP_ID);
    const allowed = customerFacingInputs(schema);
    const allowedNames = new Set(allowed.map((i) => i.name));

    // reject unknown keys
    const unknown = Object.keys(inputs).filter((k) => !allowedNames.has(k));
    if (unknown.length) {
      return res.status(422).json({ error: `unknown inputs: ${unknown.join(", ")}` });
    }

    // required + default + coerce to string
    const finalInputs = {};
    for (const def of allowed) {
      let v = inputs[def.name];
      if (v === undefined || v === "") v = def.default;
      if ((v === undefined || v === "") && def.required) {
        return res.status(422).json({ error: `missing required input: ${def.name}` });
      }
      if (v !== undefined && v !== "") finalInputs[def.name] = String(v);
    }

    const payload = {
      name: name.trim(),
      inputs: finalInputs,
      [`${CLOUD_PLATFORM}_account`]: getServerCloudAccount(customer), // TODO
      labels: { tenant_id: String(customer.tenantId) }, // TODO: your tenant id
    };

    const install = await nuon.createInstall(APP_ID, payload);
    res.status(201).json({
      id: install.id,
      name: install.name,
      sandbox_status: install.sandbox_status,
      runner_status: install.runner_status,
      created_at: install.created_at,
    });
  } catch (err) {
    mapNuonError(err, res);
  }
});

// GET /api/installs/:id — status polling (tenant-scoped)
router.get("/installs/:id", async (req, res) => {
  try {
    authorizeReadInstall(getCustomer(req), req.params.id);
    const install = await nuon.getInstall(req.params.id);
    res.json({
      id: install.id,
      name: install.name,
      sandbox_status: install.sandbox_status,
      runner_status: install.runner_status,
      composite_component_status: install.composite_component_status,
    });
  } catch (err) {
    mapNuonError(err, res);
  }
});

// TODO: return the cloud account block the install should deploy into.
// This is server-owned config, NOT customer input.
function getServerCloudAccount(/* customer */) {
  throw new Error("getServerCloudAccount not implemented");
}

module.exports = router;
