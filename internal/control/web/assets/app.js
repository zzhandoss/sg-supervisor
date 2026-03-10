const state = {
  page: "overview",
  status: null,
  logs: null,
};
let flashTimer = null;

const pageMeta = {
  overview: {
    kicker: "Overview",
    title: "System overview",
    description: "See overall status, the next recommended step, and how to open the system.",
  },
  install: {
    kicker: "Install",
    title: "First-time setup",
    description: "Prepare application files, activate the system, and complete the first launch steps.",
  },
  control: {
    kicker: "Control",
    title: "Application control",
    description: "Start, stop, and check the main services after setup is complete.",
  },
  service: {
    kicker: "Service",
    title: "Windows service host",
    description: "Manage the single Windows service that keeps Control Center and the app available after reboot.",
  },
  configuration: {
    kicker: "Configuration",
    title: "Application settings",
    description: "Manage connection settings and optional Telegram bot credentials.",
  },
  updates: {
    kicker: "Updates",
    title: "Updates",
    description: "Import signed update files and apply the version you want to run.",
  },
  maintenance: {
    kicker: "Maintenance",
    title: "Install, repair, and uninstall",
    description: "Use these actions only when changing or repairing the local installation.",
  },
};

document.addEventListener("DOMContentLoaded", () => {
  wireNavigation();
  wireCoreForms();
  wireQuickActions();
  document.querySelector('[data-action="refresh"]').addEventListener("click", () => {
    refreshStatus("Status refreshed");
  });
  renderPage();
  refreshStatus();
  window.setInterval(() => refreshStatus(), 15000);
});

function wireNavigation() {
  document.querySelectorAll("[data-page-tab]").forEach((button) => {
    button.addEventListener("click", () => {
      state.page = button.dataset.pageTab;
      renderPage();
    });
  });
}

function wireQuickActions() {
  document.addEventListener("click", async (event) => {
    const navigate = event.target.closest("[data-navigate-page]");
    if (navigate) {
      state.page = navigate.dataset.navigatePage;
      renderPage();
      return;
    }

    const bootstrap = event.target.closest("[data-bootstrap-action]");
    if (bootstrap) {
      window.__sgCurrentTrigger = bootstrap;
      try {
        await runAction(() => api("/api/v1/bootstrap/start", {}), "Application preparation started");
      } finally {
        window.__sgCurrentTrigger = null;
      }
    }
  });
}

function renderPage() {
  const meta = pageMeta[state.page] || pageMeta.overview;
  document.getElementById("page-kicker").textContent = meta.kicker;
  document.getElementById("page-title").textContent = meta.title;
  document.getElementById("page-description").textContent = meta.description;

  document.querySelectorAll("[data-page-tab]").forEach((button) => {
    button.classList.toggle("is-active", button.dataset.pageTab === state.page);
  });
  document.querySelectorAll("[data-page]").forEach((section) => {
    section.classList.toggle("is-active", section.dataset.page === state.page);
  });
}

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
  bootstrapButton.addEventListener("click", async () => {
    window.__sgCurrentTrigger = bootstrapButton;
    try {
      await runAction(() => api("/api/v1/bootstrap/start", {}), "Application preparation started");
    } finally {
      window.__sgCurrentTrigger = null;
    }
  });
}

function bindSubmit(formId, action) {
  const form = document.getElementById(formId);
  if (!form) {
    return;
  }
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    window.__sgCurrentTrigger = event.submitter || form.querySelector('button[type="submit"]');
    try {
      await runAction(async () => action(form), "Saved");
    } finally {
      window.__sgCurrentTrigger = null;
    }
  });
}

async function runAction(action, successMessage) {
  const trigger = currentTrigger();
  setBusy(trigger, true);
  try {
    const result = await action();
    flash(successMessage || "Action completed");
    await refreshStatus();
    return result;
  } catch (error) {
    flash(error.message, true);
    return null;
  } finally {
    setBusy(trigger, false);
  }
}

async function refreshStatus(successMessage) {
  try {
    const [statusResult, logsResult] = await Promise.all([
      fetchJSON("/api/v1/status"),
      fetchJSON("/api/v1/logs/recent?limit=60"),
    ]);
    state.status = statusResult.data;
    state.logs = logsResult.data;
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
  syncMaintenanceDefaults(status.root);
  renderHeaderStatus(status);
  renderOverviewStatus(status);
  renderSetupStatus(status);
  renderProductConfigStatus(status.productConfig);
  renderBootstrapStatus(status);
  renderServicesStatus(status.services || []);
  renderServiceHostStatus(status.serviceHost);
  renderUpdateStatus(status);
  renderRecentLogs(state.logs);
}

function renderHeaderStatus(status) {
  const pills = [
    pill(status.license.valid ? "valid" : "invalid", status.license.valid ? "License ok" : "License missing"),
    pill(status.bootstrap?.state || "pending", `Bootstrap ${status.bootstrap?.state || "idle"}`),
    pill(status.setupRequired ? "blocked" : "ready", status.setupRequired ? "Setup required" : "Setup complete"),
  ];
  document.getElementById("header-status-pills").innerHTML = pills.join("");
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
  const response = await fetch(path, { ...init });
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

function setFormValue(formId, fieldName, value) {
  const field = document.querySelector(`#${formId} [name="${fieldName}"]`);
  if (field) {
    field.value = value;
  }
}

function syncMaintenanceDefaults(root) {
  const binaryPath = defaultSupervisorBinaryPath(root);
  document.querySelectorAll('#install-package-form [name="binaryPath"], #repair-form [name="binaryPath"]').forEach((field) => {
    if (!field.value.trim()) {
      field.value = binaryPath;
    }
  });
}

function defaultSupervisorBinaryPath(root) {
  if (!root) {
    return "";
  }
  if (root.includes("\\")) {
    return `${root}\\sg-supervisor.exe`;
  }
  return `${root}/sg-supervisor`;
}

function selectPackageForAction(packageId) {
  setFormValue("package-apply-form", "packageId", packageId);
  setFormValue("install-package-form", "packageId", packageId);
  state.page = "updates";
  renderPage();
  flash(`Selected package ${packageId}`);
}

function renderStatusOption(current, option) {
  return `<option value="${option}" ${current === option ? "selected" : ""}>${option}</option>`;
}

function pill(tone, label) {
  return `<span class="pill ${escapeHTMLClass(tone)}">${escapeHTML(label)}</span>`;
}

function flash(message, isError) {
  const node = document.getElementById("flash-message");
  node.textContent = message;
  node.className = `flash-message is-visible ${isError ? "is-error" : "is-success"}`;
  if (flashTimer) {
    window.clearTimeout(flashTimer);
  }
  flashTimer = window.setTimeout(() => {
    node.className = "flash-message";
    node.textContent = "";
  }, 4200);
}

function currentTrigger() {
  return window.__sgCurrentTrigger || null;
}

function setBusy(trigger, busy) {
  if (!trigger) {
    return;
  }
  if (busy) {
    if (!trigger.dataset.originalLabel) {
      trigger.dataset.originalLabel = trigger.textContent;
    }
    trigger.disabled = true;
    trigger.classList.add("is-busy");
    trigger.textContent = trigger.dataset.busyLabel || "Working...";
    return;
  }
  trigger.disabled = false;
  trigger.classList.remove("is-busy");
  if (trigger.dataset.originalLabel) {
    trigger.textContent = trigger.dataset.originalLabel;
  }
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
