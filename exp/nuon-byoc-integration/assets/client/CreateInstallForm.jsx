// Create-install form. Fields are rendered dynamically from the app's input
// schema (getInstallInputs) — nothing hardcoded.

import { useEffect, useState } from "react";
import { getInstallInputs, createInstall } from "./installsApi";

function fieldControl(def, value, onChange) {
  const common = {
    id: def.name,
    value: value ?? "",
    required: def.required,
    onChange: (e) =>
      onChange(def.name, def.type === "bool" ? e.target.checked : e.target.value),
  };
  if (def.type === "bool") {
    return <input type="checkbox" checked={!!value} {...common} value={undefined} />;
  }
  if (def.type === "number") return <input type="number" {...common} />;
  if (def.sensitive) return <input type="password" {...common} />;
  return <input type="text" {...common} />;
}

export function CreateInstallForm({ onCreated }) {
  const [schema, setSchema] = useState(null);
  const [name, setName] = useState("");
  const [values, setValues] = useState({});
  const [error, setError] = useState(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    getInstallInputs()
      .then((s) => {
        setSchema(s);
        const defaults = {};
        for (const i of s.inputs) if (i.default != null) defaults[i.name] = i.default;
        setValues(defaults);
      })
      .catch((e) => setError(e.message));
  }, []);

  if (error && !schema) return <p role="alert">Failed to load: {error}</p>;
  if (!schema) return <p>Loading…</p>;

  const inputs = [...schema.inputs].sort((a, b) => (a.index ?? 0) - (b.index ?? 0));

  const setValue = (k, v) => setValues((prev) => ({ ...prev, [k]: v }));

  async function onSubmit(e) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const install = await createInstall({ name, inputs: values });
      onCreated?.(install);
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={onSubmit}>
      <label htmlFor="install-name">Install name</label>
      <input
        id="install-name"
        value={name}
        required
        onChange={(e) => setName(e.target.value)}
      />

      {inputs.map((def) => (
        <div key={def.name}>
          <label htmlFor={def.name}>{def.display_name || def.name}</label>
          {def.description && <p>{def.description}</p>}
          {fieldControl(def, values[def.name], setValue)}
        </div>
      ))}

      {error && <p role="alert">{error}</p>}
      <button type="submit" disabled={submitting}>
        {submitting ? "Creating…" : "Create"}
      </button>
    </form>
  );
}
