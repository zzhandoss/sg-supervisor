function renderOverviewStatus(status) {
  const cards = [
    { label: "Product", value: status.productName || "Unknown", meta: status.root },
    { label: "Setup", value: status.bootstrap?.state || "idle", meta: status.bootstrap?.currentStep || "waiting" },
    { label: "License", value: status.license.valid ? "Active" : "Missing", meta: status.license.customer || "not activated" },
    { label: "Services", value: `${runningServices(status.services || [])}`, meta: `${(status.services || []).length} available` },
  ];

  document.getElementById("overview-cards").innerHTML = cards
    .map(
      (card) => `
        <article class="status-card">
          <p class="status-label">${escapeHTML(card.label)}</p>
          <div class="status-card-value">${escapeHTML(card.value)}</div>
          <p class="meta-line">${escapeHTML(card.meta)}</p>
        </article>
      `
    )
    .join("");

  const nextStep = buildNextStep(status);
  document.getElementById("next-step-card").innerHTML = `
    <strong>${escapeHTML(nextStep.title)}</strong>
    <p class="summary-copy">${escapeHTML(nextStep.detail)}</p>
    ${renderActionButtons(nextStep.actions)}
  `;

  document.getElementById("access-card").innerHTML = [
    summaryLine("Panel", status.listenAddr),
    summaryLine("Preferred host", status.productConfig?.preferredHost || "auto"),
    summaryLine("API", status.productConfig?.viteApiBaseUrl || "n/a"),
    summaryLine("Admin UI", status.productConfig?.adminUiUrl || "n/a"),
  ].join("");

  document.getElementById("package-card").innerHTML = [
    summaryLine("Installed version", status.activePackage?.productVersion || "not selected"),
    summaryLine("Package id", status.activePackage?.packageId || "none"),
    summaryLine("Core version", status.activePackage?.coreVersion || "n/a"),
    summaryLine("Control Center version", status.activePackage?.supervisorVersion || "n/a"),
  ].join("");
}

function renderRecentLogs(logs) {
  const node = document.getElementById("logs-card");
  if (!node) {
    return;
  }
  if (!logs || !(logs.lines || []).length) {
    node.innerHTML = [
      summaryLine("Log file", logs?.path || "not created yet"),
      `<p class="summary-copy">No log lines yet.</p>`,
    ].join("");
    return;
  }
  node.innerHTML = [
    summaryLine("Log file", logs.path || "n/a"),
    `<div class="log-lines">${logs.lines.map((line) => `<div>${escapeHTML(line)}</div>`).join("")}</div>`,
  ].join("");
}

function renderBootstrapStatus(status) {
  const bootstrap = status.bootstrap || {};
  document.getElementById("bootstrap-summary").innerHTML = [
    summaryLine("State", bootstrap.state || "idle"),
    summaryLine("Current step", bootstrap.currentStep || "waiting"),
    summaryLine("Application files", bootstrap.sourceArchivePath || "not detected"),
    summaryLine("Adapter files", bootstrap.adapterArchivePath || "not detected"),
    summaryLine("Error", bootstrap.error || "none"),
  ].join("");
  document.getElementById("bootstrap-logs").innerHTML = renderBootstrapLogs(bootstrap.logs || []);

  const steps = bootstrap.steps || [];
  document.getElementById("bootstrap-steps").innerHTML = steps
    .map(
      (step) => `
        <article class="setup-card">
          <div class="pill-row">${pill(step.state || "pending", step.state || "pending")}</div>
          <h3>${escapeHTML(step.name)}</h3>
          <p class="meta-line">${escapeHTML(step.message || "Waiting for the previous step")}</p>
        </article>
      `
    )
    .join("");

  const stepper = buildInstallSteps(status);
  document.getElementById("install-stepper").innerHTML = stepper
    .map(
      (step, index) => `
        <article class="step-card ${step.current ? "is-current" : ""} ${step.done ? "is-done" : ""}">
          <div class="step-index">${index + 1}</div>
          <div class="pill-row">${pill(step.tone, step.status)}</div>
          <h3>${escapeHTML(step.title)}</h3>
          <p class="summary-copy">${escapeHTML(step.detail)}</p>
          ${renderActionButtons(step.actions)}
        </article>
      `
    )
    .join("");
}

function renderBootstrapLogs(lines) {
  if (!lines.length) {
    return [
      summaryLine("Recent activity", "none yet"),
      `<p class="summary-copy">The latest bootstrap messages will appear here.</p>`,
    ].join("");
  }
  const recent = lines.slice(-8);
  return [
    summaryLine("Recent activity", `${recent.length} message(s)`),
    `<div class="log-lines">${recent.map((line) => `<div>${escapeHTML(line)}</div>`).join("")}</div>`,
  ].join("");
}

function renderSetupStatus(status) {
  const setupFields = [...(status.setup.required || []), ...(status.setup.optional || [])];
  document.getElementById("license-card").innerHTML = [
    summaryLine("Status", status.license.valid ? "active" : "needs license"),
    summaryLine("Customer", status.license.customer || "not activated"),
    summaryLine("License type", status.license.mode || "n/a"),
    summaryLine("Details", status.license.reason || "ok"),
  ].join("");

  document.getElementById("setup-fields").innerHTML = setupFields
    .map((field) => renderSetupField(field))
    .join("");

  document.querySelectorAll("[data-setup-field]").forEach((form) => {
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const payload = new FormData(form);
      await runAction(
        () =>
          api("/api/v1/setup/fields", {
            key: form.dataset.setupField,
            status: payload.get("status"),
            value: String(payload.get("value") || "").trim(),
          }),
        `${form.dataset.setupField} updated`
      );
    });
  });
}

function renderSetupField(field) {
  return `
    <article class="setup-card">
      <div class="pill-row">
        ${pill(field.status, field.status)}
        ${field.required ? pill("blocked", "required") : pill("unknown", "optional")}
      </div>
      <h3>${escapeHTML(field.label)}</h3>
      <p class="summary-copy">${escapeHTML(field.required ? "Complete this before the first start." : "You can finish this later if needed.")}</p>
      <form class="form-grid" data-setup-field="${escapeHTML(field.key)}">
        <label>
          <span>Status</span>
          <select name="status">
            ${renderStatusOption(field.status, "pending")}
            ${renderStatusOption(field.status, "completed")}
            ${renderStatusOption(field.status, "skipped")}
          </select>
        </label>
        <button type="submit">Save item</button>
      </form>
    </article>
  `;
}

function renderServicesStatus(services) {
  document.getElementById("services-list").innerHTML = services
    .map(
      (service) => `
        <article class="service-card">
          <div class="service-head">
            <div>
              <h3>${escapeHTML(service.name)}</h3>
              <p class="meta-line">${escapeHTML(service.lastError || serviceMessage(service))}</p>
            </div>
            <div class="pill-row">
              ${pill(service.state, service.state)}
              ${pill(service.readiness, service.readiness)}
              ${pill(service.reachability || "unknown", service.reachability || "unknown")}
            </div>
          </div>
          <div class="service-meta">
            ${summaryLine("License", service.requiresLicense ? "required" : "not required")}
            ${summaryLine("Process", processLabel(service))}
            ${summaryLine("Local health", service.readiness || "unknown")}
            ${summaryLine("Reachable", service.reachability || "unknown")}
            ${service.primaryUrl ? summaryLine("Address", service.primaryUrl) : ""}
          </div>
          <div class="button-row button-row-compact">
            <button type="button" class="button-secondary" data-service-action="start" data-service-name="${escapeHTML(service.name)}">Start</button>
            <button type="button" class="button-secondary" data-service-action="stop" data-service-name="${escapeHTML(service.name)}">Stop</button>
            <button type="button" class="button-primary" data-service-action="restart" data-service-name="${escapeHTML(service.name)}">Restart</button>
          </div>
        </article>
      `
    )
    .join("");

  document.getElementById("control-access-card").innerHTML = [
    summaryLine("Control Panel", state.status.listenAddr),
    summaryLine("Admin UI", state.status.productConfig?.adminUiUrl || "n/a"),
    summaryLine("Preferred address", state.status.productConfig?.preferredHost || "automatic"),
  ].join("");

  document.querySelectorAll("[data-service-action]").forEach((button) => {
    button.addEventListener("click", async () => {
      const action = button.dataset.serviceAction;
      const outcome = action === "stop" ? "stopped" : `${action}ed`;
      window.__sgCurrentTrigger = button;
      try {
        await runAction(
          () => api(`/api/v1/services/${action}`, { name: button.dataset.serviceName }),
          `${button.dataset.serviceName} ${outcome}`
        );
      } finally {
        window.__sgCurrentTrigger = null;
      }
    });
  });
}

function renderServiceHostStatus(serviceHost) {
  const card = document.getElementById("service-host-card");
  if (!card || !serviceHost) {
    return;
  }

  card.innerHTML = [
    summaryLine("Service", serviceHost.serviceName || "school-gate-supervisor"),
    summaryLine("State", serviceHost.state || "unknown"),
    summaryLine("Installed", serviceHost.installed ? "yes" : "no"),
    summaryLine("Autostart", serviceHost.startMode || "unknown"),
    summaryLine("Wrapper file", serviceHost.wrapperPath || "not bundled"),
    summaryLine("Config", serviceHost.configPath || "not generated"),
    summaryLine("Details", serviceHost.lastError || serviceHost.description || "none"),
  ].join("");

  document.querySelectorAll("[data-service-host-action]").forEach((button) => {
    button.addEventListener("click", async () => {
      const action = button.dataset.serviceHostAction;
      window.__sgCurrentTrigger = button;
      try {
        await runAction(
          () => api(`/api/v1/service-host/${action}`, {}),
          serviceHostActionLabel(action)
        );
      } finally {
        window.__sgCurrentTrigger = null;
      }
    });
  });
}

function serviceHostActionLabel(action) {
  switch (action) {
    case "install":
      return "Windows service installed";
    case "switch":
      return "Switching Control Panel to Windows service mode";
    case "start":
      return "Windows service started";
    case "stop":
      return "Windows service stopped";
    case "enable-autostart":
      return "Autostart enabled";
    case "disable-autostart":
      return "Autostart disabled";
    case "remove":
      return "Windows service removed";
    default:
      return "Service host updated";
  }
}

function processLabel(service) {
  if (service.state === "running") {
    return "running";
  }
  if (service.state === "error") {
    return "failed";
  }
  return service.state || "unknown";
}

function serviceMessage(service) {
  if (service.reachability === "ready") {
    return "Service is reachable from the selected address.";
  }
  if (service.state === "running" && service.readiness === "ready") {
    return "Process is running locally. External access still needs confirmation.";
  }
  if (service.readiness === "not_ready") {
    return "Process started, but the local health check is not ready yet.";
  }
  return "This service does not expose a confirmed access check yet.";
}

function renderUpdateStatus(status) {
  const lastUpdate = status.lastUpdate || {};
  document.getElementById("update-status").innerHTML = [
    summaryLine("Last result", lastUpdate.outcome || "none"),
    summaryLine("Version", lastUpdate.packageId || "n/a"),
    summaryLine("Rollback", lastUpdate.rollbackOutcome || "n/a"),
    summaryLine("Message", lastUpdate.message || "none"),
  ].join("");

  document.getElementById("packages-list").innerHTML = (status.importedPackages || [])
    .map(
      (pkg) => `
        <article class="package-card">
          <div class="pill-row">
            ${pill("complete", pkg.packageId)}
            ${status.activePackage?.packageId === pkg.packageId ? pill("ready", "active") : ""}
          </div>
          <h3>${escapeHTML(pkg.productVersion || pkg.packageId)}</h3>
          <p class="meta-line">Loaded: ${escapeHTML(pkg.importedAt || "n/a")}</p>
          <p class="meta-line">Core: ${escapeHTML(pkg.coreVersion || "n/a")} | Control Center: ${escapeHTML(pkg.supervisorVersion || "n/a")}</p>
          <div class="button-row">
            <button type="button" class="button-secondary" data-package-select="${escapeHTML(pkg.packageId)}">Choose this version</button>
            <button type="button" class="button-primary" data-package-apply="${escapeHTML(pkg.packageId)}">Install this version</button>
          </div>
        </article>
      `
    )
    .join("");

  document.querySelectorAll("[data-package-apply]").forEach((button) => {
    button.addEventListener("click", async () => {
      window.__sgCurrentTrigger = button;
      try {
        await runAction(
          () => api("/api/v1/updates/apply", { packageId: button.dataset.packageApply }),
          `${button.dataset.packageApply} applied`
        );
      } finally {
        window.__sgCurrentTrigger = null;
      }
    });
  });

  document.querySelectorAll("[data-package-select]").forEach((button) => {
    button.addEventListener("click", () => {
      selectPackageForAction(button.dataset.packageSelect);
    });
  });
}

function summaryLine(label, value) {
  return `<div class="summary-line"><strong>${escapeHTML(label)}:</strong> ${escapeHTML(value)}</div>`;
}

function findService(services, name) {
  return (services || []).find((service) => service.name === name);
}

function runningServices(services) {
  return (services || []).filter((service) => service.state === "running").length;
}
