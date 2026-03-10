const state = {
  status: null,
  recipeDirty: false,
};
let activeButton = null;

async function api(path, options = {}) {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const payload = await response.json();
  if (!response.ok || payload.success === false) {
    throw new Error(payload.error?.message || "Request failed");
  }
  return payload.data;
}

function setBusy(button, busy, label = "Working...") {
  if (!button) {
    return;
  }
  if (busy) {
    if (!button.dataset.originalLabel) {
      button.dataset.originalLabel = button.textContent;
    }
    button.disabled = true;
    button.textContent = label;
    return;
  }
  button.disabled = false;
  if (button.dataset.originalLabel) {
    button.textContent = button.dataset.originalLabel;
  }
}

function setText(id, value) {
  document.getElementById(id).textContent = value || "";
}

function setResult(id, value, kind = "") {
  const node = document.getElementById(id);
  node.textContent = value || "";
  node.className = kind ? `result ${kind}` : "result";
}

function setRecipeStatus(message, kind = "") {
  const node = document.getElementById("recipe-status");
  node.textContent = message;
  node.className = kind ? `status-note ${kind}` : "status-note";
}

function recipeValues() {
  return {
    installerVersion: document.getElementById("installer-version").value.trim(),
    schoolGateVersion: document.getElementById("school-gate-version").value.trim(),
    adapterVersion: document.getElementById("adapter-version").value.trim(),
    nodeVersion: document.getElementById("node-version").value.trim(),
  };
}

function sameRecipe(left, right) {
  return left.installerVersion === (right.installerVersion || "")
    && left.schoolGateVersion === (right.schoolGateVersion || "")
    && left.adapterVersion === (right.adapterVersion || "")
    && left.nodeVersion === (right.nodeVersion || "");
}

function syncRecipeInputs(recipe) {
  document.getElementById("installer-version").value = recipe.installerVersion || "";
  document.getElementById("school-gate-version").value = recipe.schoolGateVersion || "";
  document.getElementById("adapter-version").value = recipe.adapterVersion || "";
  document.getElementById("node-version").value = recipe.nodeVersion || "";
}

function renderStatus(status) {
  state.status = status;
  setText("summary", `Host: ${status.hostPlatform} | Repo: ${status.repoRoot || "not set"} | Releases: ${status.releaseDir}`);
  setText("build-policy", status.hostPlatform === "windows"
    ? "This machine builds the Windows delivery archive locally. Build the Linux delivery archive from a Linux host."
    : "This machine builds the Linux delivery archive locally. Build the Windows delivery archive from a Windows host.");
  if (!state.recipeDirty) {
    syncRecipeInputs(status.recipe || {});
    setRecipeStatus("Recipe is saved and ready for local release.", "saved");
  } else {
    setRecipeStatus("You have unsaved recipe changes. Save them before starting a local release.", "pending");
  }
  renderJobs(status.jobs || []);
  renderLicenses(status.issuedLicenses || []);
  updateLicenseModeUI();
}

function renderJobs(jobs) {
  const root = document.getElementById("jobs");
  root.innerHTML = jobs.length ? "" : "<p class='meta'>No jobs yet.</p>";
  jobs.forEach((job) => {
    const article = document.createElement("article");
    article.className = "job-card";
    const logs = (job.logs || []).map((entry) => `<li>${entry}</li>`).join("");
    const artifact = job.report?.reports?.map((report) => `<li>${report.platform}: ${report.artifactPath}</li>`).join("") || "";
    article.innerHTML = `
      <div class="job-header">
        <strong>${job.type}</strong>
        <span class="badge ${job.status || ""}">${job.status}</span>
      </div>
      <span class="meta">${job.createdAt}</span>
      ${job.error ? `<span class="meta">${job.error}</span>` : ""}
      ${artifact ? `<ul class="log-list">${artifact}</ul>` : ""}
      ${logs ? `<ul class="log-list">${logs}</ul>` : ""}
    `;
    root.appendChild(article);
  });
}

function renderLicenses(records) {
  const root = document.getElementById("licenses");
  root.innerHTML = records.length ? "" : "<p class='meta'>No licenses issued yet.</p>";
  records.forEach((record) => {
    const article = document.createElement("article");
    article.className = "license-card";
    article.innerHTML = `
      <strong>${record.licenseId}</strong>
      <span class="meta">${record.customer || "No customer"} | ${record.mode} | ${record.edition}</span>
      <span class="meta">${record.path}</span>
    `;
    root.appendChild(article);
  });
}

async function refresh() {
  renderStatus(await api("/api/v1/status"));
}

async function saveRecipe(event) {
  event.preventDefault();
  activeButton = event.submitter;
  setBusy(activeButton, true, "Saving...");
  try {
    await api("/api/v1/recipe", {
      method: "POST",
      body: JSON.stringify(recipeValues()),
    });
    state.recipeDirty = false;
    setResult("recipe-result", "Recipe saved.", "success");
    setRecipeStatus("Recipe is saved and ready for local release.", "saved");
    await refresh();
  } catch (error) {
    setResult("recipe-result", error.message, "error");
  } finally {
    setBusy(activeButton, false);
    activeButton = null;
  }
}

async function fetchVersions(event) {
  const repo = event.target.dataset.repo;
  activeButton = event.target;
  setBusy(activeButton, true, "Loading...");
  try {
    const versions = await api(`/api/v1/upstream/versions?repo=${encodeURIComponent(repo)}`);
    const target = document.getElementById(`${repo}-options`);
    target.innerHTML = versions.map((entry) => `<option value="${entry.tag.replace(/^v/, "")}"></option>`).join("");
  } catch (error) {
    setResult("recipe-result", error.message, "error");
  } finally {
    setBusy(activeButton, false);
    activeButton = null;
  }
}

async function startBuild(event) {
  if (state.recipeDirty) {
    setResult("build-result", "Save the recipe before starting a local release.", "error");
    return;
  }
  activeButton = event.target;
  setBusy(activeButton, true, "Starting...");
  try {
    const job = await api("/api/v1/releases/local", { method: "POST", body: "{}" });
    setResult("build-result", `Started ${state.status?.hostPlatform || "local"} release job ${job.id}.`, "success");
    await refresh();
  } catch (error) {
    setResult("build-result", error.message, "error");
  } finally {
    setBusy(activeButton, false);
    activeButton = null;
  }
}

async function loadActivationRequestFile() {
  const file = document.getElementById("activation-request-file").files[0];
  if (!file) {
    setResult("license-result", "Choose an activation-request.json file first.", "error");
    return;
  }
  activeButton = document.getElementById("load-activation-request");
  setBusy(activeButton, true, "Loading...");
  try {
    const text = await file.text();
    const request = JSON.parse(text);
    document.getElementById("license-customer").value = request.customerHint || "";
    document.getElementById("license-fingerprint").value = request.fingerprint || "";
    document.getElementById("license-mode").value = request.fingerprint ? "bound" : "free";
    updateLicenseModeUI();
    setResult("license-result", "Activation request loaded. Customer and fingerprint fields were filled automatically.", "success");
  } catch (error) {
    setResult("license-result", error.message, "error");
  } finally {
    setBusy(activeButton, false);
    activeButton = null;
  }
}

async function issueLicense(event) {
  event.preventDefault();
  activeButton = event.submitter;
  setBusy(activeButton, true, "Issuing...");
  try {
    const mode = document.getElementById("license-mode").value;
    const perpetual = document.getElementById("license-perpetual").checked;
    const features = document.getElementById("license-features").value
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean);
    if (mode === "bound" && !document.getElementById("license-fingerprint").value.trim()) {
      throw new Error("Bound license requires a fingerprint or activation-request.json.");
    }
    if (!perpetual && !document.getElementById("license-expires-at").value.trim()) {
      throw new Error("Set Expires At or enable Perpetual before issuing the license.");
    }
    const record = await api("/api/v1/licenses/issue", {
      method: "POST",
      body: JSON.stringify({
        activationRequestPath: document.getElementById("activation-request-path").value,
        customer: document.getElementById("license-customer").value,
        mode,
        edition: document.getElementById("license-edition").value,
        features,
        expiresAt: document.getElementById("license-expires-at").value,
        fingerprint: document.getElementById("license-fingerprint").value,
        perpetual,
      }),
    });
    setResult("license-result", `Issued ${record.licenseId}. Saved to ${record.path}.`, "success");
    await refresh();
  } catch (error) {
    setResult("license-result", error.message, "error");
  } finally {
    setBusy(activeButton, false);
    activeButton = null;
  }
}

function updateLicenseModeUI() {
  const mode = document.getElementById("license-mode").value;
  const perpetualInput = document.getElementById("license-perpetual");
  if (mode === "free" && !document.getElementById("license-expires-at").value.trim()) {
    perpetualInput.checked = true;
  }
  const perpetual = perpetualInput.checked;
  const help = document.getElementById("license-mode-help");
  const bound = mode === "bound";

  document.querySelectorAll('[data-license-field="activationRequestPath"], [data-license-field="activationRequestFile"], [data-license-field="fingerprint"]').forEach((node) => {
    node.classList.toggle("is-hidden", !bound);
  });

  document.querySelector('[data-license-field="expiresAt"]').classList.toggle("is-hidden", perpetual);

  if (bound) {
    help.textContent = "Bound license mode: load activation-request.json from the customer, or paste the fingerprint manually.";
    return;
  }
  help.textContent = "Free license mode: no fingerprint is required. Choose perpetual or set an expiration date.";
}

function markRecipeDirty() {
  if (!state.status) {
    return;
  }
  state.recipeDirty = !sameRecipe(recipeValues(), state.status.recipe || {});
  if (state.recipeDirty) {
    setText("recipe-result", "");
    setRecipeStatus("You have unsaved recipe changes. Save them before starting a local release.", "pending");
    return;
  }
  setRecipeStatus("Recipe matches the saved state.", "saved");
}

document.getElementById("refresh-button").addEventListener("click", refresh);
document.getElementById("recipe-form").addEventListener("submit", saveRecipe);
document.getElementById("build-button").addEventListener("click", startBuild);
document.getElementById("license-form").addEventListener("submit", issueLicense);
document.getElementById("load-activation-request").addEventListener("click", loadActivationRequestFile);
document.getElementById("license-mode").addEventListener("change", updateLicenseModeUI);
document.getElementById("license-perpetual").addEventListener("change", updateLicenseModeUI);
document.querySelectorAll(".fetch-versions").forEach((button) => button.addEventListener("click", fetchVersions));
["installer-version", "school-gate-version", "adapter-version", "node-version"].forEach((id) => {
  document.getElementById(id).addEventListener("input", markRecipeDirty);
});

refresh().catch((error) => setText("summary", error.message));
setInterval(() => refresh().catch(() => {}), 5000);
