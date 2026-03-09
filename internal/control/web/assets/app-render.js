function renderOverviewDetails(status) {
  const nextStep = buildNextStep(status);
  const detailCards = [
    {
      title: "What to do next",
      items: [
        nextStep.title,
        nextStep.detail,
      ],
    },
    {
      title: "Access",
      items: [
        `Panel: ${status.listenAddr}`,
        `Preferred host: ${status.productConfig?.preferredHost || "auto"}`,
        `App API: ${status.productConfig?.viteApiBaseUrl || "n/a"}`,
        `Admin UI: ${status.productConfig?.adminUiUrl || "n/a"}`,
      ],
    },
    {
      title: "Current package",
      items: [
        `Package: ${status.activePackage?.packageId || "none"}`,
        `Product: ${status.activePackage?.productVersion || "n/a"}`,
        `Core: ${status.activePackage?.coreVersion || "n/a"}`,
        `Supervisor: ${status.activePackage?.supervisorVersion || "n/a"}`,
      ],
    },
  ];

  document.getElementById("status-details").innerHTML = detailCards
    .map(
      (card) => `
        <article class="detail-card">
          <p class="panel-kicker">Details</p>
          <h3>${escapeHTML(card.title)}</h3>
          <ul class="detail-list">
            ${card.items.map((item) => `<li>${escapeHTML(item)}</li>`).join("")}
          </ul>
        </article>
      `
    )
    .join("");
}

function buildNextStep(status) {
  if (!status.license.valid) {
    return {
      title: "Import a valid license file.",
      detail: "Core services stay blocked until activation is complete.",
    };
  }
  if (status.bootstrap?.state !== "succeeded") {
    return {
      title: "Run bootstrap installation from the extracted delivery archive.",
      detail: `Current bootstrap state: ${status.bootstrap?.state || "idle"}.`,
    };
  }
  if (status.setupRequired) {
    return {
      title: "Finish the required setup items.",
      detail: `Blocking fields: ${(status.setup.blockingFields || []).join(", ") || "none"}.`,
    };
  }
  if (!status.activePackage?.packageId) {
    return {
      title: "Complete bootstrap and then install or apply a package only if you need updates.",
      detail: "Fresh setup now starts from the extracted delivery archive, not from a local payload zip.",
    };
  }
  return {
    title: "Panel is ready for routine operations.",
    detail: "Use service actions, package updates, and maintenance only when needed.",
  };
}

function renderBootstrapStatus(status) {
  const summary = document.getElementById("bootstrap-summary");
  const steps = document.getElementById("bootstrap-steps");
  if (!status) {
    summary.textContent = "Bootstrap status is unavailable.";
    steps.innerHTML = "";
    return;
  }
  summary.innerHTML = `
    <strong>State:</strong> ${escapeHTML(status.state || "idle")}
    <span class="meta-line">Current step: ${escapeHTML(status.currentStep || "none")}</span>
    <span class="meta-line">Source: ${escapeHTML(status.sourceArchivePath || "not detected")}</span>
    <span class="meta-line">Adapter: ${escapeHTML(status.adapterArchivePath || "not detected")}</span>
    <span class="meta-line">Error: ${escapeHTML(status.error || "none")}</span>
  `;
  steps.innerHTML = (status.steps || []).map((step) => `
    <article class="setup-card">
      <div class="pill-row">
        ${pill(step.state || "pending", step.state || "pending")}
      </div>
      <h3>${escapeHTML(step.name)}</h3>
      <p class="meta-line">${escapeHTML(step.message || "waiting")}</p>
    </article>
  `).join("");
}

function renderSetupStatus(status) {
  const setupFields = [...(status.setup.required || []), ...(status.setup.optional || [])];
  document.getElementById("setup-fields").innerHTML =
    renderLicenseCard(status.license) + setupFields.map((field) => renderSetupField(field)).join("");

  document.querySelectorAll("[data-setup-field]").forEach((form) => {
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const field = new FormData(form);
      await runAction(
        () =>
          api("/api/v1/setup/fields", {
            key: form.dataset.setupField,
            status: field.get("status"),
            value: String(field.get("value") || "").trim(),
          }),
        `${form.dataset.setupField} updated`
      );
    });
  });
}

function renderLicenseCard(license) {
  return `
    <article class="setup-card">
      <div class="pill-row">
        ${pill(license.valid ? "valid" : "invalid", license.valid ? "valid" : "invalid")}
        ${license.mode ? pill("complete", license.mode) : ""}
      </div>
      <h3>License</h3>
      <p class="meta-line">Customer: ${escapeHTML(license.customer || "not activated")}</p>
      <p class="meta-line">Path: ${escapeHTML(license.licensePath || "missing")}</p>
      <p class="meta-line">Reason: ${escapeHTML(license.reason || "ok")}</p>
      <p class="meta-line">Expires: ${escapeHTML(license.expiresAt || "n/a")}</p>
    </article>
  `;
}

function renderSetupField(field) {
  return `
    <article class="setup-card">
      <div class="pill-row">
        ${pill(field.status, field.status)}
        ${field.required ? pill("blocked", "required") : pill("unknown", "optional")}
      </div>
      <h3>${escapeHTML(field.label)}</h3>
      <form class="stack" data-setup-field="${escapeHTML(field.key)}">
        <label>
          <span>Status</span>
          <select name="status">
            ${renderStatusOption(field.status, "pending")}
            ${renderStatusOption(field.status, "completed")}
            ${renderStatusOption(field.status, "skipped")}
          </select>
        </label>
        <button type="submit">Update field</button>
      </form>
    </article>
  `;
}

function renderServicesStatus(services) {
  document.getElementById("services-list").innerHTML = services
    .map(
      (service) => `
        <article class="service-card">
          <div class="pill-row">
            ${pill(service.state, service.state)}
            ${pill(service.readiness, service.readiness)}
            ${service.requiresLicense ? pill("blocked", "license-gated") : pill("unknown", "free")}
          </div>
          <h3>${escapeHTML(service.name)}</h3>
          <p class="meta-line">Configured: ${service.configured ? "yes" : "no"}</p>
          <p class="meta-line">License policy: ${service.requiresLicense ? "requires valid license" : "not license-gated"}</p>
          <p class="meta-line">Current note: ${escapeHTML(service.lastError || readinessMessage(service))}</p>
          <div class="button-row">
            <button type="button" data-service-action="start" data-service-name="${escapeHTML(service.name)}">Start</button>
            <button type="button" data-service-action="stop" data-service-name="${escapeHTML(service.name)}">Stop</button>
            <button type="button" data-service-action="restart" data-service-name="${escapeHTML(service.name)}">Restart</button>
          </div>
        </article>
      `
    )
    .join("");

  document.querySelectorAll("[data-service-action]").forEach((button) => {
    button.addEventListener("click", async () => {
      const action = button.dataset.serviceAction;
      const outcome = action === "stop" ? "stopped" : `${action}ed`;
      await runAction(
        () => api(`/api/v1/services/${action}`, { name: button.dataset.serviceName }),
        `${button.dataset.serviceName} ${outcome}`
      );
    });
  });
}

function readinessMessage(service) {
  if (service.readiness === "ready") {
    return "service is available";
  }
  if (service.readiness === "not_ready") {
    return "service is not ready yet";
  }
  return "readiness is not exposed for this service";
}

function renderUpdateStatus(status) {
  const lastUpdate = status.lastUpdate || {};
  document.getElementById("update-status").innerHTML = `
    <strong>Last update</strong>
    <span class="meta-line">Outcome: ${escapeHTML(lastUpdate.outcome || "none")}</span>
    <span class="meta-line">Package: ${escapeHTML(lastUpdate.packageId || "n/a")}</span>
    <span class="meta-line">Rollback: ${escapeHTML(lastUpdate.rollbackOutcome || "n/a")}</span>
    <span class="meta-line">Message: ${escapeHTML(lastUpdate.message || "none")}</span>
  `;

  document.getElementById("packages-list").innerHTML = (status.importedPackages || [])
    .map(
      (pkg) => `
        <article class="package-card">
          <div class="pill-row">
            ${pill("complete", pkg.packageId)}
            ${status.activePackage?.packageId === pkg.packageId ? pill("ready", "active") : ""}
          </div>
          <h3>${escapeHTML(pkg.productVersion || pkg.packageId)}</h3>
          <p class="meta-line">Imported: ${escapeHTML(pkg.importedAt || "n/a")}</p>
          <p class="meta-line">Source: ${escapeHTML(pkg.sourceType)} | ${escapeHTML(pkg.sourcePath)}</p>
          <p class="meta-line">Core: ${escapeHTML(pkg.coreVersion)} | Supervisor: ${escapeHTML(pkg.supervisorVersion)}</p>
          <p class="meta-line">Adapters: ${escapeHTML((pkg.adapters || []).join(", ") || "none")}</p>
          <div class="button-row">
            <button type="button" data-package-select="${escapeHTML(pkg.packageId)}">Use package id</button>
            <button type="button" data-package-apply="${escapeHTML(pkg.packageId)}">Apply imported package</button>
          </div>
        </article>
      `
    )
    .join("");

  document.querySelectorAll("[data-package-apply]").forEach((button) => {
    button.addEventListener("click", async () => {
      await runAction(
        () => api("/api/v1/updates/apply", { packageId: button.dataset.packageApply }),
        `${button.dataset.packageApply} applied`
      );
    });
  });

  document.querySelectorAll("[data-package-select]").forEach((button) => {
    button.addEventListener("click", () => {
      selectPackageForAction(button.dataset.packageSelect);
    });
  });
}
