document.addEventListener("DOMContentLoaded", () => {
  wireProductConfigForms();
});

function wireProductConfigForms() {
  bindSubmit("preferred-host-form", async (form) => {
    const result = await api("/api/v1/product-config", {
      preferredHost: form.preferredHost.value.trim(),
    });
    renderProductConfigStatus(result);
    return result;
  });

  bindSubmit("telegram-token-form", async (form) => {
    const token = form.telegramBotToken.value.trim();
    const clear = form.clearTelegramBotToken.checked;
    const payload = { clearTelegramBotToken: clear };
    if (token !== "") {
      payload.telegramBotToken = token;
    }
    const result = await api("/api/v1/product-config", payload);
    form.telegramBotToken.value = "";
    form.clearTelegramBotToken.checked = false;
    renderProductConfigStatus(result);
    return result;
  });
}

function renderProductConfigStatus(config) {
  if (!config) {
    return;
  }
  document.getElementById("product-config-summary").innerHTML = [
    summaryLine("Current address", config.resolvedHost || "n/a"),
    summaryLine("Preferred address", config.preferredHost || "automatic"),
    summaryLine("Application address", config.viteApiBaseUrl || "n/a"),
    summaryLine("Admin UI", config.adminUiUrl || "n/a"),
  ].join("");
  document.getElementById("bot-config-card").innerHTML = [
    summaryLine("Telegram bot", config.telegramBotConfigured ? "connected" : "not connected"),
    summaryLine("Current address", config.resolvedHost || "n/a"),
  ].join("");
  document.getElementById("available-hosts").innerHTML = (config.availableHosts || [])
    .map((host) => `<option value="${escapeHTML(host)}"></option>`)
    .join("");
  document.querySelector("#preferred-host-form [name=preferredHost]").value = config.preferredHost || "";
  const adminLink = document.getElementById("admin-ui-link");
  adminLink.href = config.adminUiUrl || "#";
  adminLink.setAttribute("aria-disabled", config.adminUiUrl ? "false" : "true");
}
