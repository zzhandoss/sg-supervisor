const state = {
  status: null,
  recipeDirty: false,
};

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

function setText(id, value) {
  document.getElementById(id).textContent = value || "";
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
  setText("summary", `Repo: ${status.repoRoot || "not set"} | Releases: ${status.releaseDir}`);
  if (!state.recipeDirty) {
    syncRecipeInputs(status.recipe || {});
    setRecipeStatus("Recipe is saved and ready for local release.", "saved");
  } else {
    setRecipeStatus("You have unsaved recipe changes. Save them before starting a local release.", "pending");
  }
  renderJobs(status.jobs || []);
  renderLicenses(status.issuedLicenses || []);
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
  await api("/api/v1/recipe", {
    method: "POST",
    body: JSON.stringify(recipeValues()),
  });
  state.recipeDirty = false;
  setText("recipe-result", "Recipe saved.");
  setRecipeStatus("Recipe is saved and ready for local release.", "saved");
  await refresh();
}

async function fetchVersions(event) {
  const repo = event.target.dataset.repo;
  const versions = await api(`/api/v1/upstream/versions?repo=${encodeURIComponent(repo)}`);
  const target = document.getElementById(`${repo}-options`);
  target.innerHTML = versions.map((entry) => `<option value="${entry.tag.replace(/^v/, "")}"></option>`).join("");
}

async function startBuild() {
  if (state.recipeDirty) {
    setText("build-result", "Save the recipe before starting a local release.");
    return;
  }
  const job = await api("/api/v1/releases/local", { method: "POST", body: "{}" });
  setText("build-result", `Started local release job ${job.id}.`);
  await refresh();
}

async function issueLicense(event) {
  event.preventDefault();
  const features = document.getElementById("license-features").value
    .split(",")
    .map((value) => value.trim())
    .filter(Boolean);
  const record = await api("/api/v1/licenses/issue", {
    method: "POST",
    body: JSON.stringify({
      activationRequestPath: document.getElementById("activation-request-path").value,
      customer: document.getElementById("license-customer").value,
      mode: document.getElementById("license-mode").value,
      edition: document.getElementById("license-edition").value,
      features,
      expiresAt: document.getElementById("license-expires-at").value,
      fingerprint: document.getElementById("license-fingerprint").value,
      perpetual: document.getElementById("license-perpetual").checked,
    }),
  });
  setText("license-result", `Issued ${record.licenseId}.`);
  await refresh();
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
document.querySelectorAll(".fetch-versions").forEach((button) => button.addEventListener("click", fetchVersions));
["installer-version", "school-gate-version", "adapter-version", "node-version"].forEach((id) => {
  document.getElementById(id).addEventListener("input", markRecipeDirty);
});

refresh().catch((error) => setText("summary", error.message));
setInterval(() => refresh().catch(() => {}), 5000);
