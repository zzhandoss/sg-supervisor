function buildNextStep(status) {
  if (status.bootstrap?.state !== "succeeded") {
    return {
      title: "Prepare the application files from the delivery folder.",
      detail: `Current preparation state: ${status.bootstrap?.state || "idle"}.`,
      actions: [
        { label: "Open Install", page: "install", tone: "secondary" },
        {
          label: "Prepare application",
          kind: "bootstrap",
          tone: "primary",
          disabled: status.bootstrap?.state === "running",
        },
      ],
    };
  }
  if (!status.license.valid) {
    return {
      title: "Apply a valid license file.",
      detail: "Main services stay unavailable until activation succeeds.",
      actions: [{ label: "Open activation", page: "install", tone: "primary" }],
    };
  }
  if (status.setupRequired) {
    return {
      title: "Finish the remaining setup items.",
      detail: `Pending required items: ${(status.setup.required || []).filter((item) => item.status !== "completed").map((item) => item.label).join(", ") || "none"}.`,
      actions: [{ label: "Open setup", page: "install", tone: "primary" }],
    };
  }
  if ((status.services || []).some((service) => service.name === "api" && service.state !== "running")) {
    return {
      title: "Open Control and start the main services.",
      detail: "After the first successful start, use Control for normal day-to-day actions.",
      actions: [{ label: "Open Control", page: "control", tone: "primary" }],
    };
  }
  return {
    title: "System is ready for normal work.",
    detail: "Use Configuration, Updates, and Maintenance only when you need to change something.",
    actions: [
      { label: "Open Control", page: "control", tone: "primary" },
      { label: "Open Configuration", page: "configuration", tone: "secondary" },
    ],
  };
}

function buildInstallSteps(status) {
  const bootstrapDone = status.bootstrap?.state === "succeeded";
  const bootstrapRunning = status.bootstrap?.state === "running";
  const licenseDone = status.license.valid;
  const apiRunning = findService(status.services, "api")?.state === "running";
  const botConfigured = !!status.productConfig?.telegramBotConfigured;
  return [
    {
      title: "Prepare application",
      status: bootstrapDone ? "done" : status.bootstrap?.state || "idle",
      tone: bootstrapDone ? "ready" : status.bootstrap?.state || "pending",
      detail: bootstrapDone ? "Application files were prepared successfully." : "Unpack and build the application from the delivery folder.",
      done: bootstrapDone,
      current: !bootstrapDone,
      actions: [
        { label: "Open Install", page: "install", tone: "secondary" },
        { label: "Prepare application", kind: "bootstrap", tone: "primary", disabled: bootstrapRunning || bootstrapDone },
      ],
    },
    {
      title: "Apply license",
      status: licenseDone ? "done" : "required",
      tone: licenseDone ? "ready" : "blocked",
      detail: licenseDone ? "The current license is valid." : "Apply a valid license before starting the main services.",
      done: licenseDone,
      current: bootstrapDone && !licenseDone,
      actions: [{ label: "Open activation", page: "install", tone: licenseDone ? "secondary" : "primary" }],
    },
    {
      title: "Start system",
      status: apiRunning ? "done" : "pending",
      tone: apiRunning ? "ready" : "pending",
      detail: apiRunning ? "Main services are already running." : "Open Control and start the installed system.",
      done: apiRunning,
      current: bootstrapDone && licenseDone && !apiRunning,
      actions: [{ label: "Open Control", page: "control", tone: apiRunning ? "secondary" : "primary" }],
    },
    {
      title: "Optional bot setup",
      status: botConfigured ? "configured" : "optional",
      tone: botConfigured ? "ready" : "unknown",
      detail: botConfigured ? "Telegram bot token is saved." : "You can connect Telegram later on the Configuration page.",
      done: botConfigured,
      current: false,
      actions: [{ label: "Open Configuration", page: "configuration", tone: "secondary" }],
    },
  ];
}

function renderActionButtons(actions) {
  const filtered = (actions || []).filter(Boolean);
  if (!filtered.length) {
    return "";
  }
  return `
    <div class="button-row button-row-compact">
      ${filtered
        .map((action) => {
          if (action.page) {
            return `<button type="button" class="button-${escapeHTMLClass(action.tone || "secondary")}" data-navigate-page="${escapeHTML(action.page)}">${escapeHTML(action.label)}</button>`;
          }
          if (action.kind === "bootstrap") {
            return `<button type="button" class="button-${escapeHTMLClass(action.tone || "primary")}" data-bootstrap-action="true" ${action.disabled ? "disabled" : ""}>${escapeHTML(action.label)}</button>`;
          }
          return "";
        })
        .join("")}
    </div>
  `;
}
