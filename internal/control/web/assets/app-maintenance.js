document.addEventListener("DOMContentLoaded", () => {
  wireMaintenanceForms();
});

function wireMaintenanceForms() {
  bindSubmit("install-package-form", async (form) => {
    const result = await api(
      "/api/v1/install",
      {
        packageId: form.packageId.value.trim(),
        binaryPath: form.binaryPath.value.trim(),
      },
      { allowPartial: true }
    );
    renderMaintenanceResult("Install selected version", result);
    if (!result.ok) {
      throw new Error(result.error);
    }
    return result.data;
  });

  bindSubmit("repair-form", async (form) => {
    const result = await api(
      "/api/v1/repair",
      { binaryPath: form.binaryPath.value.trim() },
      { allowPartial: true }
    );
    renderMaintenanceResult("Repair installation", result);
    if (!result.ok) {
      throw new Error(result.error);
    }
    return result.data;
  });

  bindSubmit("uninstall-form", async (form) => {
    const result = await api(
      "/api/v1/uninstall",
      { mode: form.mode.value },
      { allowPartial: true }
    );
    renderMaintenanceResult("Remove installation", result);
    if (!result.ok) {
      throw new Error(result.error);
    }
    return result.data;
  });
}

function renderMaintenanceResult(title, result) {
  const report = result.data || {};
  const issues = renderDetailItems(
    (report.issues || []).map((issue) => `${issue.step}: ${issue.message}`),
    "No issues reported"
  );
  document.getElementById("maintenance-status").innerHTML = `
    <strong>${escapeHTML(title)}</strong>
    <span class="meta-line">Result: ${result.ok ? "success" : "needs attention"}</span>
    <span class="meta-line">Finished: ${report.completed ? "yes" : "no"}</span>
    <span class="meta-line">Version: ${escapeHTML(report.activePackageId || report.packageId || "n/a")}</span>
    <span class="meta-line">Changed paths: ${String((report.removedPaths || []).length + (report.ensuredPaths || []).length)}</span>
    <ul class="issue-list">${issues}</ul>
  `;
  document.getElementById("maintenance-details").innerHTML = [
    renderMaintenanceCard("Changed files and folders", firstNonEmpty(report.removedPaths, report.ensuredPaths, report.writtenFiles), "No file changes reported"),
    renderMaintenanceCard("Other details", firstNonEmpty(report.keptPaths, report.serviceArtifacts, report.installHints), "No additional details"),
    renderMaintenanceCard("Stopped services", report.stoppedServices || [], "No services were stopped"),
  ].join("");
}

function renderMaintenanceCard(title, items, emptyLabel) {
  return `
    <article class="result-card">
      <p class="panel-kicker">Maintenance</p>
      <h3>${escapeHTML(title)}</h3>
      <ul class="detail-list">${renderDetailItems(items, emptyLabel)}</ul>
    </article>
  `;
}

function renderDetailItems(items, emptyLabel) {
  if (!items.length) {
    return `<li>${escapeHTML(emptyLabel)}</li>`;
  }
  return items.map((item) => `<li>${escapeHTML(item)}</li>`).join("");
}

function firstNonEmpty(...groups) {
  for (const group of groups) {
    if (Array.isArray(group) && group.length > 0) {
      return group;
    }
  }
  return [];
}
