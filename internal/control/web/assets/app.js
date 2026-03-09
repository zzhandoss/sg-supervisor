const state = {
  status: null,
};

document.addEventListener("DOMContentLoaded", () => {
  wireCoreForms();
  document.querySelector('[data-action="refresh"]').addEventListener("click", () => {
    refreshStatus("Status refreshed");
  });
  refreshStatus();
  window.setInterval(() => refreshStatus(), 15000);
});

function wireCoreForms() {
  bindSubmit("activation-request-form", (form) =>
    api("/api/v1/activation-request", {
      customer: form.customer.value.trim(),
      output: form.output.value.trim(),
    })
  );
  bindSubmit("license-import-form", (form) =>
    api("/api/v1/license/import", {
      path: form.path.value.trim(),
    })
  );
  bindSubmit("manifest-import-form", (form) =>
    api("/api/v1/updates/import-manifest", {
      path: form.path.value.trim(),
    })
  );
  bindSubmit("bundle-import-form", (form) =>
    api("/api/v1/updates/import-bundle", {
      path: form.path.value.trim(),
    })
  );
  bindSubmit("package-apply-form", (form) =>
    api("/api/v1/updates/apply", {
      packageId: form.packageId.value.trim(),
    })
  );
  const bootstrapButton = document.getElementById("bootstrap-start-button");
  if (bootstrapButton) {
    bootstrapButton.addEventListener("click", async () => {
      await runAction(
        () => api("/api/v1/bootstrap/start", {}),
        "Bootstrap installation started"
      );
    });
  }
}

function bindSubmit(formId, action) {
  const form = document.getElementById(formId);
  if (!form) {
    return;
  }
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runAction(async () => action(form), "Saved");
  });
}

async function runAction(action, successMessage) {
  try {
    const result = await action();
    flash(successMessage || "Action completed");
    await refreshStatus();
    return result;
  } catch (error) {
    flash(error.message, true);
    return null;
  }
}

async function refreshStatus(successMessage) {
  try {
    const result = await fetchJSON("/api/v1/status");
    state.status = result.data;
    renderStatus(state.status);
    if (successMessage) {
      flash(successMessage);
    }
  } catch (error) {
    flash(error.message, true);
  }
}

function renderStatus(status) {
  if (!status) {
    return;
  }
  const cards = [
    { label: "Product", value: status.productName || "Unknown" },
    { label: "Setup", value: status.setupRequired ? "Required" : "Complete" },
    { label: "License", value: status.license.valid ? "Valid" : "Invalid" },
    { label: "Services", value: `${(status.services || []).length}` },
  ];
  document.getElementById("overview-cards").innerHTML = cards
    .map(
      (card) => `
        <article class="card">
          <p class="panel-kicker">${escapeHTML(card.label)}</p>
          <h3>${escapeHTML(card.value)}</h3>
        </article>
      `
    )
    .join("");

  const summary = [
    `Root: ${status.root}`,
    `Listen: ${status.listenAddr}`,
    `Active package: ${status.activePackage?.packageId || "none"}`,
    `Last update: ${status.lastUpdate?.outcome || "none"}`,
    `License reason: ${status.license.reason || "ok"}`,
  ].join("\n");
  document.getElementById("status-summary").textContent = summary;
  renderOverviewDetails(status);
  renderSetupStatus(status);
  renderProductConfigStatus(status.productConfig);
  renderBootstrapStatus(status.bootstrap);
  renderServicesStatus(status.services || []);
  renderUpdateStatus(status);
}

async function api(path, payload, options = {}) {
  const result = await fetchJSON(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!result.ok && !options.allowPartial) {
    throw new Error(result.error);
  }
  return options.allowPartial ? result : result.data;
}

async function fetchJSON(path, init) {
  const response = await fetch(path, {
    ...init,
  });
  const body = await response.json();
  if (response.ok && body.success) {
    return { ok: true, data: body.data };
  }
  if (body.data) {
    return {
      ok: false,
      data: body.data,
      error: body.error?.message || `Request failed for ${path}`,
    };
  }
  throw new Error(body.error?.message || `Request failed for ${path}`);
}

function renderStatusOption(current, option) {
  return `<option value="${option}" ${current === option ? "selected" : ""}>${option}</option>`;
}

function setFormValue(formId, fieldName, value) {
  const field = document.querySelector(`#${formId} [name="${fieldName}"]`);
  if (field) {
    field.value = value;
  }
}

function selectPackageForAction(packageId) {
  setFormValue("package-apply-form", "packageId", packageId);
  setFormValue("install-package-form", "packageId", packageId);
  flash(`Selected package ${packageId}`);
}

function pill(tone, label) {
  return `<span class="pill ${escapeHTMLClass(tone)}">${escapeHTML(label)}</span>`;
}

function flash(message, isError) {
  const node = document.getElementById("flash-message");
  node.textContent = message;
  node.style.color = isError ? "var(--bad)" : "var(--ok)";
}

function escapeHTML(value) {
  return String(value || "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function escapeHTMLClass(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, "-");
}
