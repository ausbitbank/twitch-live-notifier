(() => {
  "use strict";

  const iconPaths = {
    overview: '<rect x="3" y="3" width="7" height="7" rx="2"/><rect x="14" y="3" width="7" height="7" rx="2"/><rect x="3" y="14" width="7" height="7" rx="2"/><rect x="14" y="14" width="7" height="7" rx="2"/>',
    users: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
    send: '<path d="m22 2-7 20-4-9-9-4Z"/><path d="M22 2 11 13"/>',
    settings: '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1 1.55V21a2 2 0 1 1-4 0v-.09a1.7 1.7 0 0 0-1.11-1.55 1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.55-1H3a2 2 0 1 1 0-4h.09a1.7 1.7 0 0 0 1.55-1.11 1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.7 1.7 0 0 0 9 4.6h.08a1.7 1.7 0 0 0 1-1.55V3a2 2 0 1 1 4 0v.09a1.7 1.7 0 0 0 1 1.55 1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.7 1.7 0 0 0 19.4 9v.08a1.7 1.7 0 0 0 1.55 1H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.51 1Z"/>',
    refresh: '<path d="M21 12a9 9 0 1 1-2.64-6.36"/><path d="M21 3v6h-6"/>',
    plus: '<path d="M12 5v14M5 12h14"/>',
    trash: '<path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="m19 6-1 14H6L5 6"/><path d="M10 11v5M14 11v5"/>',
    edit: '<path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z"/>',
    play: '<path d="m7 3 12 9-12 9Z"/>',
    external: '<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>',
    folder: '<path d="M3 7a2 2 0 0 1 2-2h5l2 2h7a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/>',
    power: '<path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/>',
    check: '<path d="M20 6 9 17l-5-5"/>',
    bell: '<path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"/><path d="M10 21h4"/>',
    monitor: '<rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8M12 17v4"/>',
    moon: '<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79Z"/>',
    sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.42 1.42M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/>',
    x: '<path d="M18 6 6 18M6 6l12 12"/>',
    eye: '<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="3"/>',
    link: '<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>',
    sparkles: '<path d="m12 3 1.4 4.2L18 9l-4.6 1.8L12 15l-1.4-4.2L6 9l4.6-1.8Z"/><path d="m19 15 .8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8Z"/><path d="m5 15 .7 1.8L7.5 18l-1.8.7L5 20.5l-.7-1.8L2.5 18l1.8-.7Z"/>',
    copy: '<rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>',
    download: '<path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M5 21h14"/>',
    upload: '<path d="M12 21V9"/><path d="m7 14 5-5 5 5"/><path d="M5 3h14"/>',
    history: '<path d="M3 12a9 9 0 1 0 2.64-6.36"/><path d="M3 3v6h6"/><path d="M12 7v5l3 2"/>',
    shield: '<path d="M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6Z"/><path d="m9 12 2 2 4-4"/>',
    filter: '<path d="M3 5h18l-7 8v6l-4 2v-8Z"/>',
    file: '<path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9Z"/><path d="M14 3v6h6"/>',
    info: '<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><path d="M12 8h.01"/>',
  };

  const MAX = {
    content: 2000,
    title: 256,
    description: 4096,
    footer: 2048,
  };

  const $ = (selector, root = document) => root.querySelector(selector);
  const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

  let state = null;
  let currentView = "overview";
  let refreshing = false;
  let lastFocusedField = null;
  let previewTimer = null;
  let previewSequence = 0;
  let routePreviewTimer = null;
  let routePreviewSequence = 0;
  let routeLastFocusedField = null;
  let escHandler = null;
  let twitchImportState = null;
  let twitchImportTimer = null;
  let tokenPromptActive = false;
  let tokenPromptDismissed = false;

  let importModalTab = "helper";
  let importBookmarklet = "";
  let importBookmarkletPromise = null;
  let importFileSource = false;

  let updateInstallTimer = null;
  let updateInstallEndpointMissing = false;
  let updateInstallStartedAt = 0;
  let updateInstallLastJob = null;
  let awaitingRestart = false;
  let restartPollTimer = null;
  let restartBaseVersion = "";
  let restartSawDisconnect = false;
  let updateInstallRestartRequested = false;

  let historyLoaded = false;
  let historyLoading = false;
  let historyError = "";
  let historyData = { entries: [], total: 0, limit: 200 };
  let historyFilters = { kind: "", level: "", channel_id: "", search: "" };
  let historyRequestSequence = 0;
  let historySearchTimer = null;
  let historyReloadPending = false;

  let webhookInspectTimer = null;
  let webhookInspectSequence = 0;

  let streamerSearch = "";
  let streamerStatus = "all";
  let streamerSort = "name";
  let streamerSelectMode = false;
  let streamerView = "cards";
  let selectedStreamerIds = new Set();
  let streamerSearchTimer = null;
  let bulkActionBusy = false;

  function icon(name) {
    return `<svg viewBox="0 0 24 24" aria-hidden="true">${iconPaths[name] || ""}</svg>`;
  }

  function escapeHTML(value) {
    return String(value ?? "").replace(/[&<>"']/g, (character) => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;",
    }[character]));
  }

  function safeURL(value) {
    try {
      const url = new URL(value);
      return ["http:", "https:"].includes(url.protocol) ? url.href : "";
    } catch {
      return "";
    }
  }

  function safeColor(value) {
    const normalized = String(value || "").replace("#", "");
    return /^[0-9a-fA-F]{6}$/.test(normalized) ? `#${normalized}` : "#9146FF";
  }

  function hexFromInt(value) {
    if (typeof value !== "number" || !Number.isFinite(value)) {
      return "#9146ff";
    }
    return `#${Math.max(0, value).toString(16).padStart(6, "0").slice(-6)}`;
  }

  function truncate(value, length) {
    const text = String(value || "");
    return text.length > length ? `${text.slice(0, length - 1)}…` : text;
  }

  function formatCount(value) {
    return new Intl.NumberFormat().format(Number(value) || 0);
  }

  function formatBytes(value) {
    const bytes = Number(value) || 0;
    if (bytes <= 0) {
      return "0 B";
    }
    const units = ["B", "KB", "MB", "GB"];
    const exponent = Math.min(units.length - 1, Math.floor(Math.log(bytes) / Math.log(1024)));
    const scaled = bytes / (1024 ** exponent);
    return `${scaled >= 10 || exponent === 0 ? Math.round(scaled) : scaled.toFixed(1)} ${units[exponent]}`;
  }

  function formatDateTime(value) {
    const time = new Date(value).getTime();
    if (!Number.isFinite(time)) {
      return "";
    }
    return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(time));
  }

  function formatInterval(seconds) {
    const value = Number(seconds) || 120;
    if (value < 3600) {
      return `${Math.round(value / 60)} min`;
    }
    const hours = value / 3600;
    return `${Number.isInteger(hours) ? hours : hours.toFixed(1)} hr`;
  }

  function updateIntervalOptions(current) {
    const values = [0, 60, 120, 300, 600, 900, 1800];
    const selected = Number(current) || 0;
    if (selected && !values.includes(selected)) {
      values.push(selected);
    }
    values.sort((a, b) => a - b);
    return values.map((seconds) => `<option value="${seconds}" ${seconds === selected ? "selected" : ""}>${seconds === 0 ? "Off" : `Every ${formatInterval(seconds)}`}</option>`).join("");
  }

  function formatRelative(value) {
    if (!value) {
      return "never";
    }
    const time = new Date(value).getTime();
    if (!Number.isFinite(time)) {
      return "";
    }
    const diff = Math.round((time - Date.now()) / 1000);
    const absolute = Math.abs(diff);
    const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
    if (absolute < 60) {
      return formatter.format(diff, "second");
    }
    if (absolute < 3600) {
      return formatter.format(Math.round(diff / 60), "minute");
    }
    if (absolute < 86400) {
      return formatter.format(Math.round(diff / 3600), "hour");
    }
    return formatter.format(Math.round(diff / 86400), "day");
  }

  function initials(name) {
    return String(name || "?")
      .split(/[\s_-]+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0].toUpperCase())
      .join("");
  }

  function avatarHTML(channel, extraClass = "") {
    const avatar = safeURL(channel?.avatar_url || "");
    if (avatar) {
      return `<div class="avatar ${extraClass}"><img src="${escapeHTML(avatar)}" alt=""></div>`;
    }
    return `<div class="avatar ${extraClass}">${escapeHTML(initials(channel?.display_name || channel?.login))}</div>`;
  }

  function statusFor(channelId) {
    return state?.statuses?.[channelId] || {};
  }

  function channelById(id) {
    return (state?.config?.channels || []).find((channel) => channel.id === id);
  }

  function webhookById(id) {
    return (state?.config?.webhooks || []).find((webhook) => webhook.id === id);
  }

  function assignedWebhookIDs(channel) {
    return (channel?.destinations || []).map((destination) => destination.webhook_id);
  }

  function liveChannels() {
    return (state?.config?.channels || []).filter((channel) => channel.enabled && statusFor(channel.id).live);
  }

  function webhookConnectionCount(webhookId) {
    return (state?.config?.channels || []).filter((channel) => assignedWebhookIDs(channel).includes(webhookId)).length;
  }

  function webhookRetryUntil(webhookId) {
    const now = Date.now();
    let latest = 0;
    for (const channel of state?.config?.channels || []) {
      for (const destination of channel.destinations || []) {
        if (destination.webhook_id !== webhookId) {
          continue;
        }
        const time = destination.retry_until ? new Date(destination.retry_until).getTime() : 0;
        if (Number.isFinite(time) && time > now && time > latest) {
          latest = time;
        }
      }
    }
    return latest;
  }

  function storedDashboardToken() {
    try {
      return sessionStorage.getItem("twitch-live-token") || "";
    } catch {
      return "";
    }
  }

  async function api(path, { method = "GET", body, tokenRetry = false } = {}) {
    const options = { method, headers: {}, cache: "no-store" };
    const token = storedDashboardToken();
    if (token) {
      options.headers["X-Dashboard-Token"] = token;
    }
    if (method !== "GET") {
      options.headers["Content-Type"] = "application/json";
      options.body = JSON.stringify(body ?? {});
    }
    const response = await fetch(path, options);
    if (response.status === 401 && !tokenRetry && !tokenPromptActive && !tokenPromptDismissed) {
      try {
        await response.body?.cancel();
      } catch {
        // Ignore a body that is already consumed.
      }
      tokenPromptActive = true;
      const entered = window.prompt("This dashboard requires a token. Enter it to continue.");
      tokenPromptActive = false;
      if (entered) {
        try {
          sessionStorage.setItem("twitch-live-token", entered.trim());
        } catch {
          // Ignore storage restrictions.
        }
        return api(path, { method, body, tokenRetry: true });
      }
      tokenPromptDismissed = true;
    }
    let data = {};
    try {
      data = await response.json();
    } catch {
      data = {};
    }
    if (!response.ok) {
      const error = new Error(data.error || `Request failed (${response.status})`);
      error.status = response.status;
      error.data = data;
      throw error;
    }
    return data;
  }

  function applyTheme(theme) {
    const value = theme === "light" ? "light" : "dark";
    document.documentElement.dataset.theme = value;
    try {
      localStorage.setItem("twitch-live-theme", value);
    } catch {
      // Ignore storage restrictions.
    }
  }

  function userIsTyping() {
    const element = document.activeElement;
    if (!element || !element.closest(".main")) {
      return false;
    }
    return ["INPUT", "TEXTAREA", "SELECT"].includes(element.tagName);
  }

  async function refresh(silent = false) {
    if (refreshing) {
      return;
    }
    refreshing = true;
    try {
      state = await api("/api/state");
      if (state.update && !state.update.install_job && installJobActive(updateInstallLastJob)) {
        state.update = { ...state.update, install_job: updateInstallLastJob };
      }
      applyTheme(state.config.settings.theme);
      maybeResumeUpdatePolling();
      if (userIsTyping()) {
        renderChrome();
      } else {
        render();
      }
    } catch (error) {
      if (!silent) {
        toast(error.message, "error");
      }
    } finally {
      refreshing = false;
    }
  }

  async function poll() {
    await refresh(true);
    setTimeout(poll, state?.checking ? 1200 : 4000);
  }

  function switchView(view, focusAdd = false) {
    currentView = view;
    if (location.hash !== `#${view}`) {
      history.replaceState(null, "", `#${view}`);
    }
    render();
    if (focusAdd) {
      setTimeout(() => {
        const input = $("#streamer-login");
        if (input) {
          input.focus();
          input.scrollIntoView({ behavior: "smooth", block: "center" });
        }
      }, 40);
    }
  }

  function render() {
    if (!state) {
      return;
    }
    renderChrome();
    renderHeader();
    $$(".view").forEach((view) => view.classList.add("is-hidden"));
    const viewElement = $(`#view-${currentView}`);
    if (viewElement) {
      viewElement.classList.remove("is-hidden");
    }
    $$(".nav-item").forEach((item) => item.classList.toggle("is-active", item.dataset.view === currentView));
    switch (currentView) {
      case "streamers":
        renderStreamers();
        break;
      case "destinations":
        renderDestinations();
        break;
      case "history":
        renderHistory();
        if (!historyLoaded && !historyLoading) {
          loadHistory();
        }
        break;
      case "settings":
        renderSettings();
        break;
      default:
        renderOverview();
        break;
    }
  }

  function renderChrome() {
    if (!state) {
      return;
    }
    const channels = state.config.channels || [];
    const webhooks = state.config.webhooks || [];
    $("#nav-streamer-count").textContent = String(channels.length);
    $("#nav-webhook-count").textContent = String(webhooks.length);
    $("#app-version").textContent = `v${state.version}`;

    const pill = $("#monitor-pill");
    const label = $("span:last-child", pill);
    pill.classList.remove("is-paused", "is-checking");
    let text = `Every ${formatInterval(state.config.settings.poll_seconds)}`;
    if (state.checking) {
      pill.classList.add("is-checking");
      text = "Checking…";
    } else if (state.config.settings.paused) {
      pill.classList.add("is-paused");
      text = "Paused";
    } else if (!channels.length) {
      text = "Ready to set up";
    }
    label.textContent = text;
  }

  function renderHeader() {
    const views = {
      overview: ["Overview", "Your community's live alerts at a glance."],
      streamers: ["Streamers", "Every Twitch channel you want to follow."],
      destinations: ["Destinations", "Labelled Discord webhooks and their message styles."],
      history: ["History", "Every post, edit, skip, and delivery error."],
      settings: ["Settings", "Monitoring, startup, and local data."],
    };
    const [title, subtitle] = views[currentView] || views.overview;
    $("#page-title").textContent = title;
    $("#page-subtitle").textContent = subtitle;

    const actions = [];
    if (currentView === "overview" || currentView === "streamers") {
      actions.push(`<button class="btn btn-ghost" data-action="check-all" ${state.checking ? "disabled" : ""}>${state.checking ? '<span class="spinner dark"></span>' : icon("refresh")} Check all</button>`);
      actions.push(`<button class="btn btn-primary" data-action="go-add-streamer">${icon("plus")} Add streamer</button>`);
    } else if (currentView === "destinations") {
      actions.push(`<button class="btn btn-primary" data-action="add-destination">${icon("plus")} Add destination</button>`);
    } else if (currentView === "history") {
      actions.push(`<button class="btn btn-ghost" data-action="refresh-history" ${historyLoading ? "disabled" : ""}>${historyLoading ? '<span class="spinner dark"></span>' : icon("refresh")} Refresh</button>`);
      actions.push(`<button class="btn btn-danger" data-action="clear-history">${icon("trash")} Clear history</button>`);
    }
    $("#top-actions").innerHTML = actions.join("");
  }

  function renderOverview() {
    const channels = state.config.channels || [];
    const webhooks = state.config.webhooks || [];
    const live = liveChannels();
    const connectionCount = channels.reduce((total, channel) => total + assignedWebhookIDs(channel).length, 0);

    const stats = `
      <div class="stats-grid">
        <article class="stat">
          <div class="stat-top"><span>Streamers</span>${icon("users")}</div>
          <div class="stat-value">${channels.length}</div>
          <div class="stat-note">${live.length} live right now</div>
        </article>
        <article class="stat">
          <div class="stat-top"><span>Destinations</span>${icon("send")}</div>
          <div class="stat-value">${webhooks.length}</div>
          <div class="stat-note">${connectionCount} stream connection${connectionCount === 1 ? "" : "s"}</div>
        </article>
        <article class="stat">
          <div class="stat-top"><span>Monitoring</span>${icon("monitor")}</div>
          <div class="stat-value">${state.config.settings.paused ? "Paused" : "On"}</div>
          <div class="stat-note">${state.config.settings.paused ? "Manual checks still work" : `Every ${formatInterval(state.config.settings.poll_seconds)}`}</div>
        </article>
        <article class="stat">
          <div class="stat-top"><span>Tray app</span>${icon("bell")}</div>
          <div class="stat-value">${state.platform === "windows" ? "Ready" : "Web"}</div>
          <div class="stat-note">${state.platform === "windows" ? "Keeps running when this tab closes" : "Dashboard server only"}</div>
        </article>
      </div>`;

    const hero = `
      <section class="hero">
        <div class="hero-copy">
          <div class="eyebrow">${icon("sparkles")} Community alerts</div>
          <h2>Your people never miss the moment you go live.</h2>
          <p>Add Twitch channels, connect labelled Discord webhooks, and tune the message exactly how your community likes it.</p>
          <div class="hero-actions">
            ${channels.length ? `<button class="btn btn-primary" data-action="check-all">${icon("refresh")} Check all now</button>` : `<button class="btn btn-primary" data-action="go-add-streamer">${icon("plus")} Add your first streamer</button>`}
            ${webhooks.length ? "" : `<button class="btn btn-ghost" data-action="add-destination">${icon("send")} Connect Discord</button>`}
          </div>
        </div>
        <div class="hero-art"><img src="/static/mascot.png" alt=""></div>
      </section>`;

    const onboarding = renderOnboarding(channels, webhooks);
    const liveSection = live.length
      ? `<div class="live-list">${live.map(renderLiveCard).join("")}</div>`
      : `<div class="empty-state">${icon("play")}<strong>No one is live right now</strong><span>The dashboard refreshes automatically${channels.length ? "" : " once you add a streamer"}.</span>${channels.length ? `<button class="btn btn-small btn-ghost" data-action="check-all">${icon("refresh")} Check now</button>` : `<button class="btn btn-small btn-primary" data-action="go-add-streamer">${icon("plus")} Add your first streamer</button>`}</div>`;
    const activity = renderActivity();

    $("#view-overview").innerHTML = `${stats}${hero}${onboarding}
      <div class="grid-2 section">
        <section class="card">
          <div class="card-head">
            <div><h3>Live now</h3><p>Streams currently posting to your community.</p></div>
          </div>
          <div class="card-body">${liveSection}</div>
        </section>
        <section class="card">
          <div class="card-head">
            <div><h3>Recent activity</h3><p>The latest posts, checks, and edits.</p></div>
          </div>
          <div class="card-body">${activity}</div>
        </section>
      </div>`;
  }

  function onboardingStorageKey() {
    return `twitch-live-onboarding-dismissed:${state?.version || "unknown"}`;
  }

  function onboardingDismissed() {
    try {
      return localStorage.getItem(onboardingStorageKey()) === "1";
    } catch {
      return false;
    }
  }

  function dismissOnboarding() {
    try {
      localStorage.setItem(onboardingStorageKey(), "1");
    } catch {
      // Ignore storage restrictions.
    }
    render();
  }

  function renderOnboarding(channels, webhooks) {
    if (onboardingDismissed()) {
      return "";
    }
    const routed = channels.filter((channel) => assignedWebhookIDs(channel).length > 0).length;
    const tested = webhooks.some((webhook) => Boolean(webhook.last_success_at));
    const checked = Boolean(state.last_check);
    const steps = [
      {
        done: webhooks.length > 0,
        label: "Add a Discord destination",
        hint: "Paste a webhook URL so alerts have somewhere to go.",
        action: "add-destination",
        actionLabel: "Add destination",
      },
      {
        done: channels.length > 0,
        label: "Add at least one streamer",
        hint: "Paste a Twitch username, or import the channels you already follow.",
        action: "go-add-streamer",
        actionLabel: "Add streamer",
      },
      {
        done: routed > 0,
        label: "Route a streamer to a destination",
        hint: "Choose which Discord channels receive each streamer's alerts.",
        action: "go-routing",
        actionLabel: "Review routing",
      },
      {
        done: tested,
        label: "Test a webhook or send a sample",
        hint: "A successful test confirms Discord can receive your alerts.",
        action: "go-test-webhook",
        actionLabel: "Send a test",
      },
      {
        done: checked,
        label: "Run a check",
        hint: "Checks every enabled streamer and posts any live alerts.",
        action: "check-all",
        actionLabel: "Run check",
      },
    ];
    if (steps.every((step) => step.done)) {
      return "";
    }
    const completed = steps.filter((step) => step.done).length;
    return `
      <section class="card section onboarding-card">
        <div class="card-head">
          <div><h3>Finish setup</h3><p>Five small steps and your community is covered.</p></div>
          <div class="head-actions">
            <span class="badge">${completed}/${steps.length}</span>
            <button class="btn btn-small btn-ghost" type="button" data-action="dismiss-onboarding">${icon("x")} Dismiss</button>
          </div>
        </div>
        <div class="card-body">
          <div class="step-list">
            ${steps.map((step) => `
              <div class="step ${step.done ? "is-done" : ""}">
                <span class="step-check">${step.done ? "✓" : ""}</span>
                <div class="step-copy">
                  <strong>${escapeHTML(step.label)}</strong>
                  <span>${escapeHTML(step.hint)}</span>
                </div>
                ${step.done ? "" : `<button type="button" class="btn btn-small btn-ghost" data-action="${escapeHTML(step.action)}">${escapeHTML(step.actionLabel)}</button>`}
              </div>`).join("")}
          </div>
        </div>
      </section>`;
  }

  function renderLiveCard(channel) {
    const status = statusFor(channel.id);
    const destinations = assignedWebhookIDs(channel).map(webhookById).filter(Boolean);
    return `
      <article class="live-card">
        <div class="live-top">
          ${avatarHTML(channel)}
          <div style="min-width:0;flex:1">
            <div class="meta-strong">${escapeHTML(channel.display_name || channel.login)}</div>
            <div class="meta">${escapeHTML(channel.login)}</div>
          </div>
          <span class="badge badge-live">Live</span>
        </div>
        <div class="stream-title">${escapeHTML(status.title || "Live on Twitch")}</div>
        <div class="live-stats">
          ${status.game ? `<span class="pill">${escapeHTML(status.game)}</span>` : ""}
          <span class="pill">${icon("users")} ${formatCount(status.viewers)}</span>
          ${status.started_at ? `<span class="pill">${formatRelative(status.started_at)}</span>` : ""}
        </div>
        <div class="chips">${destinations.map((webhook) => `<span class="chip"><span class="chip-dot"></span>${escapeHTML(webhook.label)}</span>`).join("") || `<span class="chip">No destinations</span>`}</div>
        <div class="channel-actions">
          <a class="btn btn-small btn-ghost" href="https://www.twitch.tv/${escapeHTML(channel.login)}" target="_blank" rel="noopener">${icon("external")} Open Twitch</a>
          ${destinations.length ? `<button class="btn btn-small btn-ghost" data-action="sample-channel" data-channel="${channel.id}">${icon("play")} Send sample</button>` : ""}
        </div>
      </article>`;
  }

  function renderActivity() {
    const activity = (state.activity || []).slice().reverse().slice(0, 10);
    if (!activity.length) {
      return `<div class="empty-state">${icon("bell")}<strong>Quiet for now</strong><span>Alerts will show up here.</span><button class="btn btn-small btn-ghost" data-action="go-history">${icon("history")} Open delivery history</button></div>`;
    }
    return `<div class="activity-list">${activity.map((item) => `
      <div class="activity-item">
        <span class="activity-dot ${escapeHTML(item.level)}"></span>
        <div class="activity-text"><p>${escapeHTML(item.message)}</p><span class="activity-time">${escapeHTML(formatRelative(item.time))}</span></div>
      </div>`).join("")}</div>`;
  }

  const HISTORY_KINDS = ["posted", "edited", "cleaned", "skipped", "error"];
  const HISTORY_LEVELS = ["info", "success", "error"];

  function historyActionLabel(kind) {
    const labels = {
      posted: "Posted",
      edited: "Updated",
      cleaned: "Cleaned up",
      skipped: "Skipped",
      error: "Failed",
    };
    return labels[String(kind || "")] || String(kind || "Event");
  }

  function historyQuery() {
    const params = new URLSearchParams();
    params.set("limit", "200");
    if (historyFilters.kind) {
      params.set("kind", historyFilters.kind);
    }
    if (historyFilters.level) {
      params.set("level", historyFilters.level);
    }
    if (historyFilters.channel_id) {
      params.set("channel_id", historyFilters.channel_id);
    }
    if (historyFilters.search.trim()) {
      params.set("q", historyFilters.search.trim());
    }
    return `/api/history?${params.toString()}`;
  }

  async function loadHistory() {
    if (historyLoading) {
      historyReloadPending = true;
      return;
    }
    historyLoading = true;
    historyError = "";
    const sequence = ++historyRequestSequence;
    if (currentView === "history" && !historyData.entries.length) {
      render();
    }
    try {
      const data = await api(historyQuery());
      if (sequence !== historyRequestSequence) {
        return;
      }
      const entries = Array.isArray(data.entries) ? data.entries : [];
      historyData = {
        entries,
        total: data.total != null && Number.isFinite(Number(data.total)) ? Number(data.total) : entries.length,
        limit: Number(data.limit) || 200,
      };
      historyLoaded = true;
    } catch (error) {
      if (sequence !== historyRequestSequence) {
        return;
      }
      historyError = error.message || "Could not load delivery history.";
      historyLoaded = true;
    } finally {
      historyLoading = false;
      if (historyReloadPending) {
        historyReloadPending = false;
        loadHistory();
        return;
      }
      if (currentView === "history") {
        render();
      }
    }
  }

  function refreshHistory() {
    historyLoaded = false;
    loadHistory();
  }

  function filteredHistoryEntries() {
    const entries = historyData.entries || [];
    const term = historyFilters.search.trim().toLowerCase();
    if (!term) {
      return entries;
    }
    return entries.filter((entry) => [
      entry.channel_name,
      entry.channel_login,
      entry.webhook_label,
      entry.title,
      entry.game,
      entry.reason,
      entry.kind,
      entry.level,
    ].join(" ").toLowerCase().includes(term));
  }

  function renderHistoryRow(entry) {
    const kind = String(entry.kind || "event");
    const level = String(entry.level || "info");
    const channel = entry.channel_name || entry.channel_login || "System";
    const label = entry.webhook_label || "";
    const messageURL = safeURL(entry.message_url || "");
    const time = entry.time || "";
    const detail = entry.title || entry.game
      ? `<div class="history-stream">${escapeHTML(truncate(entry.title || entry.game, 160))}${entry.title && entry.game ? ` · ${escapeHTML(entry.game)}` : ""}</div>`
      : "";
    return `
      <article class="history-row">
        <div class="history-when">
          <time datetime="${escapeHTML(time)}" title="${escapeHTML(formatDateTime(time))}">${escapeHTML(formatRelative(time))}</time>
          <span class="history-kind ${escapeHTML(kind)}">${escapeHTML(historyActionLabel(kind))}</span>
        </div>
        <div class="history-main">
          <div class="history-line">
            <strong>${escapeHTML(channel)}</strong>
            ${label ? `<span class="history-arrow" aria-hidden="true">→</span><span class="history-destination">${escapeHTML(label)}</span>` : ""}
            ${messageURL ? `<a class="history-link" href="${escapeHTML(messageURL)}" target="_blank" rel="noopener">${icon("external")} Open in Discord</a>` : ""}
          </div>
          ${detail}
          ${entry.reason ? `<div class="history-reason">${escapeHTML(entry.reason)}</div>` : ""}
        </div>
        <span class="badge history-level ${escapeHTML(level)}">${escapeHTML(level)}</span>
      </article>`;
  }

  function renderHistory() {
    const channels = state.config.channels || [];
    const entries = filteredHistoryEntries();
    const total = Number(historyData.total) || 0;
    const searchActive = historyFilters.search.trim() !== "";
    const filtersChanged = Boolean(historyFilters.kind || historyFilters.level || historyFilters.channel_id);
    const kindOptions = HISTORY_KINDS.map((kind) => `<option value="${kind}" ${historyFilters.kind === kind ? "selected" : ""}>${escapeHTML(historyActionLabel(kind))}</option>`).join("");
    const levelOptions = HISTORY_LEVELS.map((level) => `<option value="${level}" ${historyFilters.level === level ? "selected" : ""}>${escapeHTML(level)}</option>`).join("");
    const channelOptions = channels.map((channel) => `<option value="${escapeHTML(channel.id)}" ${historyFilters.channel_id === channel.id ? "selected" : ""}>${escapeHTML(channel.display_name || channel.login)}</option>`).join("");

    let content = "";
    if (historyError) {
      content = `<div class="warning-box">${icon("bell")} <span>${escapeHTML(historyError)}</span></div>
        <div class="history-empty-actions"><button class="btn btn-small btn-ghost" data-action="refresh-history">${icon("refresh")} Try again</button></div>`;
    } else if (historyLoading && !entries.length) {
      content = `<div class="empty-state">${icon("history")}<strong>Loading delivery history…</strong><span>Reading locally stored records.</span></div>`;
    } else if (!entries.length) {
      content = searchActive || filtersChanged
        ? `<div class="empty-state">${icon("filter")}<strong>No matching deliveries</strong><span>Clear a filter or search a different streamer.</span><button class="btn btn-small btn-ghost" data-action="reset-history-filters">${icon("refresh")} Clear filters</button></div>`
        : `<div class="empty-state">${icon("history")}<strong>No deliveries yet</strong><span>Posts, edits, skips, and errors appear here after the next check.</span><button class="btn btn-small btn-ghost" data-action="check-all">${icon("refresh")} Run a check</button></div>`;
    } else {
      content = `<div class="history-list">${entries.map(renderHistoryRow).join("")}</div>`;
    }

    $("#view-history").innerHTML = `
      <section class="card section">
        <div class="card-head">
          <div><h3>Delivery history</h3><p>Stored locally. Up to 1,000 recent posts, edits, skips, and errors.</p></div>
          <span class="badge">${formatCount(searchActive || filtersChanged ? entries.length : total)} record${(searchActive || filtersChanged ? entries.length : total) === 1 ? "" : "s"}</span>
        </div>
        <div class="card-body">
          <div class="history-toolbar">
            <select id="history-kind" class="select" aria-label="Filter by action">
              <option value="">All actions</option>
              ${kindOptions}
            </select>
            <select id="history-level" class="select" aria-label="Filter by level">
              <option value="">All levels</option>
              ${levelOptions}
            </select>
            <select id="history-channel" class="select" aria-label="Filter by streamer">
              <option value="">All streamers</option>
              ${channelOptions}
            </select>
            <input id="history-search" class="input" value="${escapeHTML(historyFilters.search)}" placeholder="Search streamer, destination, or reason" aria-label="Search delivery history" autocomplete="off">
          </div>
          ${content}
        </div>
      </section>`;
  }

  async function clearHistory() {
    if (!window.confirm("Clear all delivery history? This cannot be undone. Streamers, destinations, and settings are not affected.")) {
      return;
    }
    try {
      await api("/api/history", { method: "DELETE", body: {} });
      historyData = { entries: [], total: 0, limit: 200 };
      historyLoaded = true;
      historyError = "";
      render();
      toast("Delivery history cleared.", "info");
    } catch (error) {
      toast(error.message, "error");
    }
  }

  async function fetchDiagnostics() {
    return api("/api/diagnostics");
  }

  async function copyDiagnostics() {
    try {
      const data = await fetchDiagnostics();
      const text = JSON.stringify(data, null, 2);
      await navigator.clipboard.writeText(text);
      toast("Redacted diagnostics copied to the clipboard.", "success");
    } catch (error) {
      toast(error?.message || "Could not copy diagnostics. Try Download diagnostics instead.", "error");
    }
  }

  async function downloadDiagnostics() {
    try {
      const data = await fetchDiagnostics();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      const date = new Date().toISOString().slice(0, 10);
      link.href = url;
      link.download = `twitch-live-diagnostics-${date}.json`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      toast("Diagnostics downloaded.", "success");
    } catch (error) {
      toast(error?.message || "Could not download diagnostics.", "error");
    }
  }

  const STREAMER_STATUS_OPTIONS = [
    ["all", "All statuses"],
    ["live", "Live now"],
    ["offline", "Offline"],
    ["paused", "Paused"],
    ["error", "Has errors"],
    ["no-destinations", "No destinations"],
  ];

  const STREAMER_SORT_OPTIONS = [
    ["name", "Name A–Z"],
    ["live", "Live first"],
    ["recent", "Recently checked"],
  ];

  function streamerName(channel) {
    return String(channel?.display_name || channel?.login || channel?.id || "Unknown");
  }

  function streamerIsLive(channel) {
    return Boolean(channel?.enabled && statusFor(channel?.id).live);
  }

  function streamerMatchesStatus(channel) {
    const status = statusFor(channel?.id);
    switch (streamerStatus) {
      case "live":
        return streamerIsLive(channel);
      case "offline":
        return Boolean(channel?.enabled) && !status.live && !status.error;
      case "paused":
        return !channel?.enabled;
      case "error":
        return Boolean(status.error);
      case "no-destinations":
        return assignedWebhookIDs(channel).length === 0;
      default:
        return true;
    }
  }

  function streamerLastChecked(channel) {
    const time = new Date(statusFor(channel?.id).last_checked || "").getTime();
    return Number.isFinite(time) ? time : 0;
  }

  function filteredStreamerChannels() {
    const term = streamerSearch.trim().toLowerCase();
    const channels = (state?.config?.channels || []).filter((channel) => {
      if (!streamerMatchesStatus(channel)) {
        return false;
      }
      if (!term) {
        return true;
      }
      return `${channel.login || ""} ${channel.display_name || ""}`.toLowerCase().includes(term);
    });
    channels.sort((a, b) => {
      if (streamerSort === "live") {
        const liveDiff = Number(streamerIsLive(b)) - Number(streamerIsLive(a));
        if (liveDiff) {
          return liveDiff;
        }
      } else if (streamerSort === "recent") {
        const checkedDiff = streamerLastChecked(b) - streamerLastChecked(a);
        if (checkedDiff) {
          return checkedDiff;
        }
      }
      return streamerName(a).localeCompare(streamerName(b), undefined, { sensitivity: "base" });
    });
    return channels;
  }

  function channelBadgeHTML(channel) {
    if (!channel?.enabled) {
      return '<span class="badge badge-paused">Paused</span>';
    }
    return statusFor(channel.id).live
      ? '<span class="badge badge-live">Live</span>'
      : '<span class="badge badge-offline">Offline</span>';
  }

  function syncStreamerSelection() {
    if (!selectedStreamerIds.size) {
      return;
    }
    const known = new Set((state?.config?.channels || []).map((channel) => channel.id));
    for (const id of Array.from(selectedStreamerIds)) {
      if (!known.has(id)) {
        selectedStreamerIds.delete(id);
      }
    }
  }

  function syncStreamerToolbarState() {
    const selectButton = $('[data-action="toggle-select-mode"]');
    if (selectButton) {
      selectButton.classList.toggle("is-active", streamerSelectMode);
      selectButton.setAttribute("aria-pressed", streamerSelectMode ? "true" : "false");
    }
    const viewButton = $('[data-action="toggle-streamer-view"]');
    if (viewButton) {
      const matrix = streamerView === "matrix";
      viewButton.classList.toggle("is-active", matrix);
      viewButton.setAttribute("aria-pressed", matrix ? "true" : "false");
    }
  }

  function renderStreamers() {
    syncStreamerSelection();
    const channels = state.config.channels || [];
    const statusOptions = STREAMER_STATUS_OPTIONS.map(([value, label]) => `<option value="${value}" ${streamerStatus === value ? "selected" : ""}>${escapeHTML(label)}</option>`).join("");
    const sortOptions = STREAMER_SORT_OPTIONS.map(([value, label]) => `<option value="${value}" ${streamerSort === value ? "selected" : ""}>${escapeHTML(label)}</option>`).join("");
    const importButton = `<button class="btn btn-small btn-ghost" type="button" data-action="open-import">${icon("download")} Import follows</button>`;
    const addForm = `
      <section class="card section">
        <div class="card-head">
          <div><h3>Add a streamer</h3><p>Paste a Twitch username or a full channel URL.</p></div>
          <div class="head-actions">
            ${importButton}
          </div>
        </div>
        <div class="card-body">
          <form id="add-streamer-form" class="form-row" autocomplete="off">
            <input id="streamer-login" class="input" name="login" placeholder="twitch.tv/your_channel" aria-label="Twitch username or URL">
            <button class="btn btn-primary" type="submit" style="flex:none">${icon("plus")} Add streamer</button>
          </form>
          <p class="field-hint" style="margin:9px 0 0">No limit. Add every channel your community cares about.</p>
        </div>
      </section>`;

    // The list section and toolbar are stable hooks for phase-2 bulk controls.
    $("#view-streamers").innerHTML = `${addForm}<section class="section" id="streamer-list-section">
      <div class="section-head">
        <div><h2>Your streamers</h2><p id="streamer-count-line">${formatCount(channels.length)} configured</p></div>
        <div class="head-actions" id="streamer-bulk-toolbar" data-phase="bulk">
          <button type="button" class="btn btn-small btn-ghost toolbar-toggle ${streamerSelectMode ? "is-active" : ""}" data-action="toggle-select-mode" aria-pressed="${streamerSelectMode ? "true" : "false"}">${icon("check")} Select</button>
          <button type="button" class="btn btn-small btn-ghost toolbar-toggle ${streamerView === "matrix" ? "is-active" : ""}" data-action="toggle-streamer-view" aria-pressed="${streamerView === "matrix" ? "true" : "false"}">${icon("send")} Routing</button>
        </div>
      </div>
      <div class="streamer-controls">
        <input id="streamer-search" class="input" type="search" value="${escapeHTML(streamerSearch)}" placeholder="Search by name or login" aria-label="Search streamers" autocomplete="off" spellcheck="false">
        <select id="streamer-status" class="select" aria-label="Filter streamers by status">${statusOptions}</select>
        <select id="streamer-sort" class="select" aria-label="Sort streamers">${sortOptions}</select>
      </div>
      <div id="streamer-action-bar"></div>
      <div id="streamer-results"></div>
    </section>`;
    renderStreamerResults();
  }

  function streamerEmptyState() {
    return `<div class="empty-state section">${icon("users")}<strong>No streamers yet</strong><span>Add one Twitch channel by hand, or import the channels you already follow — no Twitch sign-in needed.</span><button class="btn btn-small btn-primary" data-action="open-import">${icon("download")} Import follows</button></div>`;
  }

  function renderStreamerActionBar(visible) {
    const bar = $("#streamer-action-bar");
    if (!bar) {
      return;
    }
    if (!streamerSelectMode && !selectedStreamerIds.size) {
      bar.innerHTML = "";
      return;
    }
    const count = selectedStreamerIds.size;
    const disabled = count ? "" : " disabled";
    bar.innerHTML = `
      <div class="selection-bar" role="region" aria-label="Bulk streamer actions" aria-live="polite">
        <div class="selection-count"><strong>${formatCount(count)}</strong><span>selected</span></div>
        <div class="selection-links">
          <button type="button" class="btn btn-small btn-ghost" data-action="select-visible-streamers"${visible.length ? "" : " disabled"}>${icon("check")} Select all visible</button>
          <button type="button" class="btn btn-small btn-ghost" data-action="clear-streamer-selection"${disabled}>${icon("x")} Clear</button>
        </div>
        <div class="selection-actions">
          <button type="button" class="btn btn-small btn-ghost" data-action="bulk-enable"${disabled}>${icon("check")} Enable</button>
          <button type="button" class="btn btn-small btn-ghost" data-action="bulk-disable"${disabled}>${icon("power")} Disable</button>
          <button type="button" class="btn btn-small btn-ghost" data-action="bulk-assign"${disabled}>${icon("send")} Assign destinations</button>
          <button type="button" class="btn btn-small btn-ghost" data-action="bulk-unassign"${disabled}>${icon("x")} Unassign destinations</button>
          <button type="button" class="btn btn-small btn-danger" data-action="bulk-delete"${disabled}>${icon("trash")} Delete</button>
        </div>
      </div>`;
  }

  function renderStreamerResults() {
    const results = $("#streamer-results");
    if (!results) {
      return;
    }
    syncStreamerSelection();
    const channels = state?.config?.channels || [];
    const webhooks = state?.config?.webhooks || [];
    const visible = filteredStreamerChannels();
    const countLine = $("#streamer-count-line");
    if (countLine) {
      const filtered = visible.length !== channels.length;
      countLine.textContent = `${formatCount(channels.length)} configured${filtered ? ` · ${formatCount(visible.length)} shown` : ""}`;
    }
    renderStreamerActionBar(visible);

    if (!channels.length) {
      results.innerHTML = streamerEmptyState();
      return;
    }
    if (!visible.length) {
      results.innerHTML = `<div class="empty-state">${icon("filter")}<strong>No streamers match</strong><span>Try a different search or status filter.</span><button class="btn btn-small btn-ghost" data-action="reset-streamer-filters">${icon("refresh")} Clear search & filters</button></div>`;
      return;
    }
    if (streamerView === "matrix") {
      results.innerHTML = renderRoutingMatrix(visible, webhooks);
      return;
    }
    results.innerHTML = visible.map((channel) => renderChannelCard(channel, webhooks)).join("");
  }

  function renderRoutingMatrix(channels, webhooks) {
    if (!webhooks.length) {
      return `<div class="empty-state section">${icon("send")}<strong>No destinations to route to</strong><span>Connect a Discord webhook first, then choose which streamers post there.</span><button class="btn btn-small btn-primary" data-action="add-destination">${icon("plus")} Add destination</button></div>`;
    }
    const header = webhooks.map((webhook) => {
      const stateLabel = webhook.invalid ? "Invalid" : webhook.enabled ? "Active" : "Paused";
      return `<th scope="col" class="matrix-dest-head">
        <button type="button" class="matrix-dest-link" data-action="edit-webhook" data-webhook="${escapeHTML(webhook.id)}" title="Edit destination">${escapeHTML(webhook.label)}</button>
        <span class="matrix-dest-meta">${escapeHTML(stateLabel)}</span>
      </th>`;
    }).join("");
    const rows = channels.map((channel) => {
      const assigned = new Set(assignedWebhookIDs(channel));
      const name = streamerName(channel);
      const selected = selectedStreamerIds.has(channel.id);
      const selectBox = streamerSelectMode
        ? `<input type="checkbox" class="matrix-select" data-action="toggle-select-channel" data-channel="${escapeHTML(channel.id)}" ${selected ? "checked" : ""} aria-label="${escapeHTML(`Select ${name} for bulk actions`)}">`
        : "";
      const cells = webhooks.map((webhook) => {
        const label = `Post ${name} to ${webhook.label || "destination"}`;
        return `<td class="matrix-cell"><label class="matrix-check" title="${escapeHTML(label)}"><input type="checkbox" data-action="toggle-matrix-cell" data-channel="${escapeHTML(channel.id)}" data-webhook="${escapeHTML(webhook.id)}" ${assigned.has(webhook.id) ? "checked" : ""} aria-label="${escapeHTML(label)}"></label></td>`;
      }).join("");
      return `<tr class="${selected ? "is-selected" : ""}">
        <th scope="row" class="matrix-row-head">
          <div class="matrix-identity">
            ${selectBox}
            ${avatarHTML(channel, "avatar-sm")}
            <div class="matrix-name">
              <strong>${escapeHTML(name)}</strong>
              <span>twitch.tv/${escapeHTML(channel.login || "")}</span>
            </div>
            ${channelBadgeHTML(channel)}
            <a class="matrix-row-link" href="https://www.twitch.tv/${escapeHTML(channel.login || "")}" target="_blank" rel="noopener" title="Open on Twitch" aria-label="${escapeHTML(`Open ${name} on Twitch`)}">${icon("external")}</a>
          </div>
        </th>
        ${cells}
      </tr>`;
    }).join("");
    return `<div class="matrix-scroll" tabindex="0" role="region" aria-label="Streamer and destination routing matrix">
        <table class="matrix-table">
          <thead><tr><th scope="col" class="matrix-corner">Streamer</th>${header}</tr></thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
      <p class="field-hint matrix-hint">Toggle a cell to start or stop posting that streamer's alerts to the destination. Changes save immediately.</p>`;
  }

  function renderChannelCard(channel, webhooks) {
    const status = statusFor(channel.id);
    const assigned = new Set(assignedWebhookIDs(channel));
    const destinationsByWebhook = new Map((channel.destinations || []).map((destination) => [destination.webhook_id, destination]));
    const badge = channelBadgeHTML(channel);
    const selected = selectedStreamerIds.has(channel.id);

    const destinationControls = webhooks.length
      ? `<div class="destination-checks">${webhooks.map((webhook) => {
          const destination = destinationsByWebhook.get(webhook.id);
          const customTemplate = Boolean(destination?.template);
          const retryAt = destination?.retry_until ? new Date(destination.retry_until).getTime() : 0;
          const rateLimited = Number.isFinite(retryAt) && retryAt > Date.now();
          return `<div class="destination-check-row">
            <label class="destination-check ${webhook.enabled ? "" : "is-disabled"}">
              <input type="checkbox" data-action="toggle-destination" data-channel="${channel.id}" data-webhook="${webhook.id}" ${assigned.has(webhook.id) ? "checked" : ""}>
              <span>${escapeHTML(webhook.label)}</span>
              ${webhook.enabled ? "" : '<span class="badge badge-disabled">Paused</span>'}
              ${rateLimited ? '<span class="badge badge-warn">Rate limited</span>' : ""}
            </label>
            ${assigned.has(webhook.id) ? `<button type="button" class="btn btn-small btn-ghost" data-action="edit-route-template" data-channel="${channel.id}" data-webhook="${webhook.id}">${icon("edit")} ${customTemplate ? "Custom template" : "Template"}</button>` : ""}
          </div>`;
        }).join("")}</div>`
      : `<div class="warning-box">${icon("send")} Add a Discord destination first, then choose it here.</div>`;

    const liveBlock = status.live
      ? `<div class="channel-live">
          ${status.thumbnail ? `<div class="channel-thumb"><img src="${escapeHTML(safeURL(status.thumbnail))}" alt=""></div>` : ""}
          <div style="min-width:0">
            <div class="stream-title">${escapeHTML(status.title || "Live on Twitch")}</div>
            <div class="meta" style="margin-top:4px">${escapeHTML(status.game || "No category")} · ${formatCount(status.viewers)} viewers</div>
          </div>
        </div>`
      : status.last_title
        ? `<div class="meta">Last stream: <span class="meta-strong">${escapeHTML(status.last_title)}</span></div>`
        : "";

    return `
      <article class="channel-card${selected ? " is-selected" : ""}">
        <div class="channel-head">
          ${streamerSelectMode ? `<label class="card-select" title="Select for bulk actions"><input type="checkbox" data-action="toggle-select-channel" data-channel="${escapeHTML(channel.id)}" ${selected ? "checked" : ""} aria-label="${escapeHTML(`Select ${streamerName(channel)} for bulk actions`)}"></label>` : ""}
          ${avatarHTML(channel, "avatar-lg")}
          <div class="channel-identity">
            <strong>${escapeHTML(channel.display_name || channel.login)}</strong>
            <span>twitch.tv/${escapeHTML(channel.login)}</span>
          </div>
          ${badge}
          <div class="channel-actions">
            <button class="btn btn-small btn-ghost" data-action="check-channel" data-channel="${channel.id}">${icon("refresh")} Check now</button>
            ${assigned.size ? `<button class="btn btn-small btn-ghost" data-action="sample-channel" data-channel="${channel.id}">${icon("play")} Send sample</button>` : ""}
            <button class="btn btn-small btn-ghost" data-action="toggle-channel" data-channel="${channel.id}">${channel.enabled ? "Pause" : "Resume"}</button>
            <button class="btn btn-small btn-danger" data-action="remove-channel" data-channel="${channel.id}">${icon("trash")} Remove</button>
          </div>
        </div>
        <div class="channel-body">
          ${liveBlock}
          ${status.error ? `<div class="channel-error">${escapeHTML(status.error)}</div>` : ""}
          ${status.live && status.filter_reason ? `<div class="warning-box">${icon("bell")}<span>Not posting: ${escapeHTML(status.filter_reason)}</span></div>` : ""}
          <div>
            <div style="display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:7px">
              <div class="field-label">Alert filters</div>
              <button class="btn btn-small btn-ghost" data-action="edit-filters" data-channel="${channel.id}">${icon("settings")} Edit filters</button>
            </div>
            <div class="chips">${filterChips(channel.filters).map((chip) => `<span class="chip">${escapeHTML(chip)}</span>`).join("") || `<span class="meta">No filters — every live stream alerts.</span>`}</div>
          </div>
          <div>
            <div class="field-label" style="margin-bottom:7px">Post to</div>
            ${destinationControls}
          </div>
          <div>
            <div class="field-label" style="margin-bottom:7px">Discord ping</div>
            <div class="meta" style="display:flex;align-items:center;gap:9px;flex-wrap:wrap">
              ${channel.discord_id ? `<code class="discord-id">${escapeHTML(channel.discord_id)}</code>` : "<span>Not set</span>"}
              <button class="btn btn-small btn-ghost" data-action="edit-discord-id" data-channel="${channel.id}">${icon("edit")} ${channel.discord_id ? "Edit" : "Add"}</button>
            </div>
          </div>
          <div class="meta">Last checked ${escapeHTML(formatRelative(status.last_checked))}</div>
        </div>
      </article>`;
  }

  function selectedStreamerChannels() {
    return Array.from(selectedStreamerIds).map((id) => channelById(id)).filter(Boolean);
  }

  function toggleStreamerSelection(input) {
    if (!input?.dataset.channel) {
      return;
    }
    if (input.checked) {
      selectedStreamerIds.add(input.dataset.channel);
    } else {
      selectedStreamerIds.delete(input.dataset.channel);
    }
    input.closest(".channel-card, tr")?.classList.toggle("is-selected", input.checked);
    renderStreamerActionBar(filteredStreamerChannels());
  }

  function clearStreamerSelection() {
    selectedStreamerIds.clear();
    $$('#streamer-results input[data-action="toggle-select-channel"]').forEach((input) => {
      input.checked = false;
    });
    $$("#streamer-results .is-selected").forEach((element) => element.classList.remove("is-selected"));
    renderStreamerActionBar(filteredStreamerChannels());
  }

  function selectVisibleStreamers() {
    for (const channel of filteredStreamerChannels()) {
      selectedStreamerIds.add(channel.id);
    }
    renderStreamerResults();
  }

  function resetStreamerFilters() {
    streamerSearch = "";
    streamerStatus = "all";
    streamerSort = "name";
    renderStreamers();
  }

  function openBulkDestinationModal(action) {
    const assign = action === "assign";
    const selected = selectedStreamerChannels();
    if (!selected.length) {
      toast("Select at least one streamer first.", "error");
      return;
    }
    const webhooks = state?.config?.webhooks || [];
    if (!webhooks.length) {
      toast("Add a Discord destination first, then route streamers to it.", "info");
      switchView("destinations");
      return;
    }
    const options = webhooks.map((webhook) => {
      const connected = selected.filter((channel) => assignedWebhookIDs(channel).includes(webhook.id)).length;
      const checked = assign ? webhooks.length === 1 : connected > 0;
      return `<label class="destination-check">
        <input type="checkbox" name="bulk-webhook" value="${escapeHTML(webhook.id)}" ${checked ? "checked" : ""}>
        <span>${escapeHTML(webhook.label)}</span>
        ${webhook.enabled ? "" : '<span class="badge badge-disabled">Paused</span>'}
        ${connected ? `<span class="destination-count">${connected}/${selected.length}</span>` : ""}
      </label>`;
    }).join("");
    openModal(`
      <div class="modal-head">
        <div>
          <h2>${assign ? "Assign destinations" : "Unassign destinations"}</h2>
          <p>${formatCount(selected.length)} streamer${selected.length === 1 ? "" : "s"} selected. ${assign ? "Alerts will also post to every destination you check." : "Alerts will stop posting to every destination you check."}</p>
        </div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="bulk-route-form" data-bulk-action="${assign ? "assign" : "unassign"}">
        <div class="modal-body">
          <div class="destination-checks bulk-route-list">${options}</div>
          <p class="field-hint" style="margin-top:10px">${assign ? "A destination already assigned is left unchanged." : "Only the checked destinations are removed; other routing stays."}</p>
        </div>
        <div class="modal-foot">
          <span class="field-hint">${assign ? "Assign" : "Unassign"} destinations for the selected streamers.</span>
          <div class="modal-foot-right">
            <button type="button" class="btn btn-ghost" data-action="close-modal">Cancel</button>
            <button type="submit" class="btn ${assign ? "btn-primary" : "btn-danger"}">${icon(assign ? "send" : "x")} ${assign ? "Assign destinations" : "Unassign destinations"}</button>
          </div>
        </div>
      </form>`, { narrow: true });
  }

  function bulkActionResultMessage(action, response, fallbackCount, destinationCount) {
    const updated = Number(response?.updated) || 0;
    const removed = Number(response?.removed) || 0;
    const count = (action === "delete" ? removed : updated) || fallbackCount;
    const suffix = count === 1 ? "" : "s";
    const destinations = formatCount(destinationCount || 0);
    switch (action) {
      case "enable":
        return `Enabled ${count} streamer${suffix}.`;
      case "disable":
        return `Disabled ${count} streamer${suffix}.`;
      case "delete":
        return `Removed ${count} streamer${suffix}.`;
      case "assign":
        return `Assigned ${count} streamer${suffix} to ${destinations} destination${Number(destinationCount) === 1 ? "" : "s"}.`;
      case "unassign":
        return `Unassigned ${count} streamer${suffix} from ${destinations} destination${Number(destinationCount) === 1 ? "" : "s"}.`;
      default:
        return "Streamers updated.";
    }
  }

  async function runBulkAction(action, webhookIDs = []) {
    if (bulkActionBusy) {
      return;
    }
    syncStreamerSelection();
    const ids = Array.from(selectedStreamerIds).filter((id) => Boolean(channelById(id)));
    if (!ids.length) {
      toast("Select at least one streamer first.", "error");
      return;
    }
    if (action === "delete") {
      const names = ids.slice(0, 4).map((id) => streamerName(channelById(id))).filter(Boolean);
      const extra = ids.length > names.length ? ` and ${ids.length - names.length} more` : "";
      const named = names.length ? ` (${names.join(", ")}${extra})` : "";
      if (!window.confirm(`Remove ${ids.length} streamer${ids.length === 1 ? "" : "s"}${named}? This cannot be undone. Destinations stay, but their routing and stream status are deleted.`)) {
        return;
      }
    }
    bulkActionBusy = true;
    try {
      const response = await api("/api/channels/bulk-action", {
        method: "POST",
        body: { channel_ids: ids, action, webhook_ids: webhookIDs },
      });
      if (response?.state) {
        state = response.state;
      } else {
        await refresh(true);
      }
      if (action === "delete") {
        selectedStreamerIds.clear();
      } else {
        syncStreamerSelection();
      }
      if (action === "assign" || action === "unassign") {
        closeModal();
      }
      render();
      toast(bulkActionResultMessage(action, response, ids.length, webhookIDs.length), action === "delete" ? "info" : "success");
    } catch (error) {
      toast(error.message, "error");
    } finally {
      bulkActionBusy = false;
    }
  }

  async function toggleMatrixCell(input) {
    const channel = channelById(input?.dataset.channel);
    const webhookID = input?.dataset.webhook;
    if (!channel || !webhookID) {
      if (input) {
        input.checked = !input.checked;
      }
      return;
    }
    const webhookIDs = new Set(assignedWebhookIDs(channel));
    if (input.checked) {
      webhookIDs.add(webhookID);
    } else {
      webhookIDs.delete(webhookID);
    }
    input.disabled = true;
    try {
      const response = await api(`/api/channels/${encodeURIComponent(channel.id)}/destinations`, {
        method: "PUT",
        body: { webhook_ids: Array.from(webhookIDs) },
      });
      if (response?.state) {
        state = response.state;
      } else {
        // Defensive fallback for older builds that return only a status body.
        const target = channelById(channel.id);
        if (target) {
          const existing = new Map((target.destinations || []).map((destination) => [destination.webhook_id, destination]));
          target.destinations = Array.from(webhookIDs).map((id) => existing.get(id) || { webhook_id: id });
        }
      }
      const saved = channelById(channel.id);
      if (saved) {
        input.checked = new Set(assignedWebhookIDs(saved)).has(webhookID);
      }
    } catch (error) {
      input.checked = !input.checked;
      toast(error.message, "error");
    } finally {
      input.disabled = false;
    }
  }

  function renderDestinations() {
    const webhooks = state.config.webhooks || [];
    const cards = webhooks.length
      ? webhooks.map((webhook) => renderDestinationCard(webhook)).join("")
      : `<div class="empty-state">${icon("send")}<strong>No Discord destinations yet</strong><span>Add a webhook for each Discord channel or server you want to post into.</span><button class="btn btn-small btn-primary" data-action="add-destination">${icon("plus")} Add destination</button></div>`;
    $("#view-destinations").innerHTML = `
      <div class="section-head">
        <div><h2>Discord destinations</h2><p>Each destination can have its own label, webhook, and message style.</p></div>
      </div>
      ${cards}`;
  }

  function renderDestinationCard(webhook) {
    const connected = webhookConnectionCount(webhook.id);
    const template = webhook.template || state.default_template;
    const summary = truncate(template.content || template.embed_title || template.embed_description || "No message text", 120);
    const retryAt = webhookRetryUntil(webhook.id);
    const rateLimited = retryAt > 0;
    const badge = webhook.invalid
      ? '<span class="badge badge-invalid">Invalid</span>'
      : rateLimited
        ? '<span class="badge badge-warn">Rate limited</span>'
        : `<span class="badge ${webhook.enabled ? "badge-enabled" : "badge-disabled"}">${webhook.enabled ? "Active" : "Paused"}</span>`;
    const rateLimitLine = rateLimited
      ? `<div class="meta" style="margin-top:6px">Discord asked us to wait; retrying ${escapeHTML(formatRelative(new Date(retryAt).toISOString()))}.</div>`
      : "";
    const error = webhook.last_error
      ? `<div class="${webhook.invalid ? "channel-error" : "warning-box"}" style="margin-top:10px">${icon("bell")} <span>${escapeHTML(webhook.last_error)}${webhook.last_error_at ? ` · ${escapeHTML(formatRelative(webhook.last_error_at))}` : ""}</span></div>`
      : "";
    const lastSuccess = webhook.last_success_at
      ? ` · last success ${formatRelative(webhook.last_success_at)}`
      : "";
    const lifecycle = [];
    if (webhook.update_seconds > 0) {
      lifecycle.push(`Live updates every ${formatInterval(webhook.update_seconds)}`);
    }
    if (webhook.cleanup?.mode) {
      const labels = {
        delete: "delete the live post",
        edit: "edit the live post",
        post: "post an ended message",
      };
      lifecycle.push(`Stream end: ${labels[webhook.cleanup.mode] || webhook.cleanup.mode}`);
    }
    const lifecycleLine = lifecycle.length
      ? `<div class="meta" style="margin-top:6px">${lifecycle.map((item) => escapeHTML(item)).join(" · ")}</div>`
      : "";
    return `
      <article class="destination-card">
        <div class="destination-head">
          <div class="destination-icon">${icon("send")}</div>
          <div class="destination-copy">
            <strong>${escapeHTML(webhook.label)}</strong>
            <code>${escapeHTML(maskWebhook(webhook.url))}</code>
          </div>
          ${badge}
        </div>
        <div class="template-preview-line"><strong>Message:</strong> ${escapeHTML(summary)}</div>
        ${(webhook.role_ids || []).length ? `<div class="meta" style="margin-top:6px">Role pings: ${(webhook.role_ids || []).map((id) => `<code class="discord-id">${escapeHTML(id)}</code>`).join(" ")}</div>` : ""}
        ${lifecycleLine}
        ${rateLimitLine}
        <div class="meta" style="margin-top:8px">${connected} streamer${connected === 1 ? "" : "s"} connected${lastSuccess}</div>
        ${error}
        <div class="destination-actions">
          <button class="btn btn-small btn-ghost" data-action="edit-webhook" data-webhook="${webhook.id}">${icon("edit")} Edit & preview</button>
          <button class="btn btn-small btn-ghost" data-action="test-webhook" data-webhook="${webhook.id}">${icon("send")} ${webhook.invalid ? "Retry / test" : "Test webhook"}</button>
          <button class="btn btn-small btn-danger" data-action="remove-webhook" data-webhook="${webhook.id}">${icon("trash")} Remove</button>
        </div>
      </article>`;
  }

  const INSTALL_ACTIVE_STATUSES = ["downloading", "verifying", "applying", "restarting"];

  function installJobActive(job) {
    return Boolean(job && INSTALL_ACTIVE_STATUSES.includes(String(job.status || "")));
  }

  function installJobPercent(job) {
    if (!job) {
      return 0;
    }
    const total = Number(job.total) || 0;
    const bytes = Number(job.bytes) || 0;
    let percent = 0;
    if (total > 0) {
      percent = (bytes / total) * 100;
    } else {
      const progress = Number(job.progress) || 0;
      percent = progress > 0 && progress <= 1 ? progress * 100 : progress;
    }
    if (job.status === "done") {
      percent = 100;
    }
    return Math.max(0, Math.min(100, Math.round(percent)));
  }

  function installJobLabel(job) {
    const labels = {
      downloading: "Downloading update",
      verifying: "Verifying checksum",
      applying: "Installing update",
      restarting: "Restarting…",
      done: "Update installed",
      error: "Update failed",
    };
    return labels[String(job?.status || "")] || "Update";
  }

  function installModeLabel(mode) {
    const labels = {
      in_place: "In place",
      app_bundle: "macOS app bundle",
      appimage: "AppImage",
      guided: "Guided / manual",
      unsupported: "Manual download",
    };
    return labels[String(mode || "")] || (mode ? String(mode) : "Not reported");
  }

  function renderInstallJob(job) {
    if (!job || !job.status || job.status === "idle") {
      return "";
    }
    const percent = installJobPercent(job);
    const bytes = Number(job.bytes) || 0;
    const total = Number(job.total) || 0;
    const size = total > 0 ? `${formatBytes(bytes)} of ${formatBytes(total)}` : (bytes > 0 ? formatBytes(bytes) : "");
    const statusIcon = job.status === "done"
      ? icon("check")
      : job.status === "error"
        ? icon("bell")
        : '<span class="spinner dark"></span>';
    const detail = job.status === "error" ? (job.error || job.message || "") : (job.message || "");
    const restarting = job.status === "restarting";
    return `
      <div class="install-job ${escapeHTML(job.status)}">
        <div class="install-job-head">
          <span class="install-job-state">${statusIcon} ${escapeHTML(installJobLabel(job))}</span>
          <span class="meta">${percent}%</span>
        </div>
        <div class="progress" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow="${percent}" aria-label="Update progress">
          <span class="progress-bar" style="width:${percent}%"></span>
        </div>
        <div class="install-job-meta">
          <span>${escapeHTML(size)}</span>
          ${detail ? `<span>${escapeHTML(detail)}</span>` : ""}
        </div>
        ${restarting ? `<div class="restart-note"><span class="spinner dark"></span><span>The app is restarting. This page reconnects automatically and shows the new version.</span></div>` : ""}
      </div>`;
  }

  function renderUpdateStatus(update) {
    const current = escapeHTML(update.current_version || state.version);
    const latest = update.latest_version ? escapeHTML(update.latest_version) : "";
    const assetName = update.asset_name ? escapeHTML(update.asset_name) : "";
    const assetSize = Number(update.asset_size) || 0;
    const mode = installModeLabel(update.install_mode);
    const reason = update.install_reason ? escapeHTML(update.install_reason) : "";
    const job = update.install_job || null;

    let html = "";
    if (update.disabled) {
      html = `<div class="meta">Update checks are off. You are on version <strong>${current}</strong>.</div>`;
    } else if (update.error) {
      html = `<div class="warning-box">${icon("bell")} <span>Could not check for updates: ${escapeHTML(update.error)}</span></div>`;
    } else if (update.update_available) {
      html = `<div class="warning-box">${icon("download")} <span>Version <strong>${latest}</strong> is available. You are on <strong>${current}</strong>. ${update.can_self_update ? "The app verifies the download checksum, installs it, and restarts. Your settings are preserved." : "This install needs a manual download; your settings are preserved."}</span></div>`;
    } else if (update.latest_version) {
      html = `<div class="meta">You are on the latest version, <strong>${current}</strong>${update.checked_at ? ` · checked ${escapeHTML(formatRelative(update.checked_at))}` : ""}.</div>`;
    } else {
      html = `<div class="meta">You are on version <strong>${current}</strong>. No update check has run yet.</div>`;
    }

    const facts = [
      ["Current version", current],
      ["Latest version", latest || "Not checked"],
      ["Release asset", assetName ? `${assetName}${assetSize ? ` · ${formatBytes(assetSize)}` : ""}` : "Not reported"],
      ["Install method", `${escapeHTML(mode)}${reason ? ` — ${reason}` : ""}`],
    ];
    html += `<div class="update-facts">
      ${facts.map(([label, value]) => `<div class="update-fact"><span>${escapeHTML(label)}</span><strong>${value}</strong></div>`).join("")}
    </div>`;
    html += renderInstallJob(job);
    if (update.update_available && update.can_self_update === false && update.install_reason) {
      html += `<div class="field-hint" style="margin-top:10px">Manual fallback: download the release asset, quit the app, and replace the existing file. Your settings live in a separate data folder.</div>`;
    }
    return { html };
  }

  function renderUpdateActions(update) {
    const activeJob = installJobActive(update.install_job);
    const installable = Boolean(update.update_available && update.can_self_update);
    const disabled = update.disabled ? "disabled" : "";
    return `
      <button class="btn btn-ghost" data-action="check-updates" ${disabled || activeJob ? "disabled" : ""}>${icon("refresh")} Check now</button>
      ${installable
        ? `<button class="btn btn-primary" data-action="install-update" ${activeJob ? "disabled" : ""}>${activeJob ? '<span class="spinner"></span> Installing…' : `${icon("download")} Install ${escapeHTML(update.latest_version || "update")}`}</button>`
        : update.update_available && update.download_url
          ? `<a class="btn btn-primary" href="${escapeHTML(safeURL(update.download_url))}" target="_blank" rel="noopener">${icon("download")} Download ${escapeHTML(update.latest_version || "update")}</a>`
          : ""}
      ${update.release_url ? `<a class="btn btn-ghost" href="${escapeHTML(safeURL(update.release_url))}" target="_blank" rel="noopener">${icon("external")} Release notes</a>` : ""}`;
  }

  function updateInstallUI() {
    if (!state || currentView !== "settings") {
      return;
    }
    const body = $("#update-card-body");
    const actions = $("#update-card-actions");
    if (body && actions) {
      const update = state.update || { current_version: state.version };
      body.innerHTML = renderUpdateStatus(update).html;
      actions.innerHTML = renderUpdateActions(update);
    } else if (!userIsTyping()) {
      render();
    }
  }

  function scheduleUpdateInstallPoll(delay = 700) {
    clearTimeout(updateInstallTimer);
    updateInstallTimer = setTimeout(pollUpdateInstall, delay);
  }

  function maybeResumeUpdatePolling() {
    const job = state?.update?.install_job;
    if (!job || !job.status) {
      return;
    }
    if (job.status === "restarting") {
      updateInstallRestartRequested = true;
      beginRestartWatch();
      return;
    }
    if (installJobActive(job) && !updateInstallTimer && !awaitingRestart) {
      updateInstallRestartRequested = true;
      updateInstallStartedAt = updateInstallStartedAt || Date.now();
      scheduleUpdateInstallPoll(700);
    }
  }

  function beginRestartWatch() {
    if (awaitingRestart) {
      return;
    }
    awaitingRestart = true;
    restartSawDisconnect = false;
    restartBaseVersion = state?.version || "";
    if (!updateInstallStartedAt) {
      updateInstallStartedAt = Date.now();
    }
    clearTimeout(updateInstallTimer);
    updateInstallTimer = null;
    if (state) {
      const currentJob = state.update?.install_job || {};
      state.update = {
        ...(state.update || {}),
        install_job: { ...currentJob, status: "restarting", message: currentJob.message || "Restarting the app…" },
      };
    }
    updateInstallUI();
    clearTimeout(restartPollTimer);
    restartPollTimer = setTimeout(checkRestarted, 1200);
  }

  async function checkRestarted() {
    restartPollTimer = null;
    const startedAt = updateInstallStartedAt || Date.now();
    updateInstallStartedAt = startedAt;
    try {
      const data = await api("/api/state");
      const versionChanged = Boolean(restartBaseVersion && data.version && data.version !== restartBaseVersion);
      if (!versionChanged && !restartSawDisconnect && Date.now() - startedAt < 180000) {
        restartPollTimer = setTimeout(checkRestarted, 1500);
        return;
      }
      state = data;
      applyTheme(state.config.settings.theme);
      awaitingRestart = false;
      updateInstallStartedAt = 0;
      updateInstallLastJob = null;
      updateInstallRestartRequested = false;
      restartBaseVersion = "";
      restartSawDisconnect = false;
      render();
      toast(`Updated to version ${state.version}.`, "success");
    } catch {
      restartSawDisconnect = true;
      if (Date.now() - startedAt < 180000) {
        restartPollTimer = setTimeout(checkRestarted, 1500);
      } else {
        awaitingRestart = false;
        toast("The app is taking a while to restart. Reload this page when it is back.", "info");
      }
    }
  }

  function applyInstallJob(job) {
    updateInstallLastJob = job && job.status && job.status !== "idle" ? job : null;
    if (!state) {
      return;
    }
    state.update = {
      ...(state.update || { current_version: state.version }),
      install_job: updateInstallLastJob,
    };
    updateInstallUI();
  }

  async function pollUpdateInstall() {
    updateInstallTimer = null;
    if (!state) {
      return;
    }
    let job = null;
    try {
      if (!updateInstallEndpointMissing) {
        try {
          job = await api("/api/updates/install");
        } catch (error) {
          if (error.status === 404 || error.status === 405) {
            updateInstallEndpointMissing = true;
          } else if (installJobActive(updateInstallLastJob)) {
            beginRestartWatch();
            return;
          } else {
            throw error;
          }
        }
      }
      if (updateInstallEndpointMissing) {
        const update = await api("/api/updates");
        state.update = update || state.update;
        job = update?.install_job || { status: "idle" };
      }
      if (job) {
        applyInstallJob(job);
      }
      if (job && job.status === "restarting") {
        beginRestartWatch();
        return;
      }
      if (installJobActive(job)) {
        scheduleUpdateInstallPoll(700);
        return;
      }
      if (job && job.status === "done") {
        if (updateInstallRestartRequested) {
          // A successful restart-enabled install finishes with "done" and then
          // the old process exits. Wait for the dashboard to come back.
          applyInstallJob(job);
          beginRestartWatch();
          return;
        }
        toast("Update installed. Refreshing version details.", "success");
        setTimeout(() => refresh(true), 600);
      }
    } catch (error) {
      if (installJobActive(updateInstallLastJob)) {
        beginRestartWatch();
      } else {
        toast(`Could not read update progress: ${error.message}`, "error");
      }
    }
  }

  async function installUpdate() {
    updateInstallStartedAt = Date.now();
    awaitingRestart = false;
    updateInstallRestartRequested = true;
    const startButton = $("#update-card-actions [data-action='install-update']");
    if (startButton) {
      startButton.disabled = true;
      startButton.innerHTML = '<span class="spinner"></span> Starting…';
    }
    try {
      const response = await api("/api/updates/install", { method: "POST", body: { restart: true } });
      const job = response?.install_job || response?.job || (response?.status ? response : null);
      if (response && response.current_version && !response.install_job) {
        state.update = response;
      }
      if (job) {
        applyInstallJob(job);
      }
      scheduleUpdateInstallPoll(500);
      toast("Downloading and verifying the update…", "info");
    } catch (error) {
      if (error.status === 409) {
        toast("An update is already in progress. Showing its progress.", "info");
        scheduleUpdateInstallPoll(300);
        return;
      }
      toast(error.message, "error");
      updateInstallUI();
    }
  }

  function renderSettings() {
    const settings = state.config.settings;
    const update = state.update || { current_version: state.version };
    const updateStatus = renderUpdateStatus(update);
    const intervals = [60, 120, 300, 600, 900, 1800];
    if (!intervals.includes(settings.poll_seconds)) {
      intervals.push(settings.poll_seconds);
      intervals.sort((a, b) => a - b);
    }
    const intervalOptions = intervals.map((seconds) => `<option value="${seconds}" ${seconds === settings.poll_seconds ? "selected" : ""}>${formatInterval(seconds)}</option>`).join("");
    const startAtLogin = state.start_at_login_supported
      ? `<div class="switch-row">
          <div class="switch-copy"><strong>Start with Windows</strong><span>Launch quietly in the tray when you sign in.</span></div>
          <label class="switch"><input id="settings-startup" type="checkbox" ${settings.start_at_login ? "checked" : ""}><span class="switch-track"></span></label>
        </div>`
      : "";
    const windowsTools = state.platform === "windows"
      ? `<button class="btn btn-ghost" data-action="open-data-folder">${icon("folder")} Open data folder</button>`
      : "";
    const token = storedDashboardToken();
    const tokenQuery = token ? `?token=${encodeURIComponent(token)}` : "";

    $("#view-settings").innerHTML = `
      <div class="settings-grid section">
        <section class="card">
          <div class="card-head"><div><h3>Monitoring</h3><p>Choose how often the app checks Twitch.</p></div></div>
          <div class="card-body">
            <form id="settings-form">
              <div class="form-grid two">
                <label class="field">
                  <span class="field-label">Check interval</span>
                  <select id="settings-interval" class="select">${intervalOptions}</select>
                  <span class="field-hint">Twitch is checked for every enabled streamer.</span>
                </label>
                <label class="field">
                  <span class="field-label">Theme</span>
                  <select id="settings-theme" class="select">
                    <option value="dark" ${settings.theme === "dark" ? "selected" : ""}>Dark</option>
                    <option value="light" ${settings.theme === "light" ? "selected" : ""}>Light</option>
                  </select>
                  <span class="field-hint">Dark is the default.</span>
                </label>
              </div>
              <div class="settings-section">
                <div class="switch-row">
                  <div class="switch-copy"><strong>Pause monitoring</strong><span>Manual checks still work while paused.</span></div>
                  <label class="switch"><input id="settings-paused" type="checkbox" ${settings.paused ? "checked" : ""}><span class="switch-track"></span></label>
                </div>
                ${startAtLogin}
                <div class="switch-row">
                  <div class="switch-copy"><strong>Check GitHub for updates</strong><span>Reports new releases. Installs only start when you click Install, and the checksum is verified first.</span></div>
                  <label class="switch"><input id="settings-update-check" type="checkbox" ${settings.disable_update_check ? "" : "checked"}><span class="switch-track"></span></label>
                </div>
              </div>
              <div class="settings-actions">
                <button class="btn btn-primary" type="submit">${icon("check")} Save settings</button>
              </div>
            </form>
          </div>
        </section>
        <div>
          <section class="card">
            <div class="card-head"><div><h3>Local data</h3><p>Your settings never leave this computer.</p></div></div>
            <div class="card-body">
              <div class="data-path">${escapeHTML(state.data_dir)}</div>
              <p class="field-hint" style="margin:9px 0 0">Webhook URLs are moved into an encrypted secret store where your OS keychain is available, with a private 0600 key file as a fallback. Backups redact them by default.</p>
              <div class="settings-actions">
                <a class="btn btn-ghost" href="/api/config/export${tokenQuery}" download>${icon("download")} Download backup</a>
                <button class="btn btn-ghost" data-action="export-secrets">${icon("download")} With secrets</button>
                <button class="btn btn-ghost" data-action="import-config">${icon("upload")} Restore backup</button>
                <input id="config-import" type="file" accept="application/json,.json" style="display:none">
              </div>
              <div class="settings-actions">${windowsTools}<button class="btn btn-danger" data-action="quit-app">${icon("power")} Quit app</button></div>
            </div>
          </section>
          <section class="card" style="margin-top:16px">
            <div class="card-head"><div><h3>Updates</h3><p>Verified installs from GitHub releases.</p></div></div>
            <div class="card-body">
              <div class="update-status" id="update-card-body">${updateStatus.html}</div>
              <div class="settings-actions" id="update-card-actions">${renderUpdateActions(update)}</div>
            </div>
          </section>
          <section class="card" style="margin-top:16px">
            <div class="card-head"><div><h3>Dashboard access</h3><p>Optional token for dashboards exposed beyond loopback.</p></div></div>
            <div class="card-body">
              <form id="dashboard-token-form">
                <label class="field">
                  <span class="field-label">Dashboard token</span>
                  <input id="dashboard-token" class="input input-mono" type="password" placeholder="${state.dashboard_token_set ? "Enter a new token" : "At least 4 characters"}" autocomplete="new-password">
                  <span class="field-hint">${state.dashboard_token_set ? "A token is required for API changes. Enter a new one to replace it." : "No token required. The dashboard stays on loopback by default."}</span>
                </label>
                <div class="settings-actions">
                  <button class="btn btn-primary" type="submit">${icon("check")} Save token</button>
                  ${state.dashboard_token_set ? `<button class="btn btn-danger" data-action="clear-dashboard-token">${icon("trash")} Remove token</button>` : ""}
                </div>
              </form>
            </div>
          </section>
          <section class="card" style="margin-top:16px">
            <div class="card-head"><div><h3>About</h3><p>Community alerts without the busywork.</p></div></div>
            <div class="card-body">
              <div class="about-mascot">
                <img src="/static/icon.png" alt="">
                <div><strong>Twitch Live to Discord</strong><p>Version ${escapeHTML(state.version)} · Local only</p></div>
              </div>
              <p class="field-hint" style="margin:14px 0 0">Not affiliated with Twitch or Discord. Twitch and Discord are trademarks of their respective owners.</p>
            </div>
          </section>
          <section class="card" style="margin-top:16px">
            <div class="card-head"><div><h3>Diagnostics</h3><p>Redacted app health for troubleshooting.</p></div></div>
            <div class="card-body">
              <p class="field-hint" style="margin:0">The bundle includes versions, platform, channel and destination counts, recent errors, and delivery history. Webhook URLs, tokens, and other secrets are never included.</p>
              <div class="settings-actions">
                <button class="btn btn-ghost" data-action="copy-diagnostics">${icon("copy")} Copy diagnostics</button>
                <button class="btn btn-ghost" data-action="download-diagnostics">${icon("download")} Download diagnostics</button>
                <button class="btn btn-ghost" data-action="go-history">${icon("history")} Open history</button>
              </div>
            </div>
          </section>
        </div>
      </div>`;
  }

  function maskWebhook(raw) {
    try {
      const url = new URL(raw);
      const parts = url.pathname.split("/").filter(Boolean);
      const id = parts[2] || "…";
      const token = parts[3] || "";
      return `discord.com/api/webhooks/${id}/••••${token.slice(-4)}`;
    } catch {
      return raw;
    }
  }

  function parseRoleIDs(value) {
    return String(value || "")
      .split(/[\s,]+/)
      .map((part) => part.trim())
      .filter(Boolean)
      .map((part) => part.replace(/^<@&/, "").replace(/>$/, ""))
      .filter(Boolean);
  }

  function roleIDsText(roleIDs) {
    return (roleIDs || []).join(", ");
  }

  function parseCustomFields(value) {
    return String(value || "")
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => {
        const parts = line.split("|").map((part) => part.trim());
        return {
          name: parts[0] || "",
          value: parts[1] || "",
          inline: /^(inline|yes|true|1)$/i.test(parts[2] || ""),
        };
      })
      .filter((field) => field.name || field.value);
  }

  function customFieldsText(fields) {
    return (fields || [])
      .map((field) => `${field.name || ""} | ${field.value || ""}${field.inline ? " | inline" : ""}`)
      .join("\n");
  }

  function parseFieldOrder(value) {
    return String(value || "")
      .split(/[\s,]+/)
      .map((part) => part.trim())
      .filter(Boolean);
  }

  function fieldOrderText(order) {
    return (order || []).join(", ");
  }

  function defaultFieldOrder(template) {
    const builtIns = [];
    if (template?.show_game) {
      builtIns.push("game");
    }
    if (template?.show_viewers) {
      builtIns.push("viewers");
    }
    if (template?.show_live_since) {
      builtIns.push("live");
    }
    return builtIns.join(", ");
  }

  function parseButtons(value) {
    return String(value || "")
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => {
        const separator = line.indexOf("|");
        if (separator < 0) {
          return { label: line, url: "" };
        }
        return { label: line.slice(0, separator).trim(), url: line.slice(separator + 1).trim() };
      })
      .filter((button) => button.label || button.url);
  }

  function buttonsText(buttons) {
    return (buttons || []).map((button) => `${button.label || ""} | ${button.url || ""}`).join("\n");
  }

  function templatePingStreamer(template) {
    return template?.ping_streamer !== false;
  }

  function templatePingRoles(template) {
    return template?.ping_roles !== false;
  }

  const DAY_NAMES = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

  function listFrom(value) {
    return String(value || "")
      .split(/[\n,]+/)
      .map((part) => part.trim())
      .filter(Boolean);
  }

  function filterChips(filters) {
    const f = filters || {};
    const chips = [];
    if ((f.include_games || []).length) {
      chips.push(`Game: ${f.include_games.join(", ")}`);
    }
    if ((f.exclude_games || []).length) {
      chips.push(`Not game: ${f.exclude_games.join(", ")}`);
    }
    if ((f.include_titles || []).length) {
      chips.push(`Title: ${f.include_titles.join(", ")}`);
    }
    if ((f.exclude_titles || []).length) {
      chips.push(`Not title: ${f.exclude_titles.join(", ")}`);
    }
    if (f.min_viewers > 0) {
      chips.push(`Min ${formatCount(f.min_viewers)} viewers`);
    }
    if (f.delay_minutes > 0) {
      chips.push(`After ${f.delay_minutes} min`);
    }
    if (f.quiet_start && f.quiet_end) {
      chips.push(`Quiet ${f.quiet_start}–${f.quiet_end}`);
    }
    if ((f.allowed_days || []).length) {
      chips.push(`Days: ${f.allowed_days.map((day) => DAY_NAMES[day] || day).join(", ")}`);
    }
    if (f.cooldown_minutes > 0) {
      chips.push(`Cooldown ${f.cooldown_minutes} min`);
    }
    if (f.first_stream_of_day) {
      chips.push("First stream/day");
    }
    return chips;
  }


  function openModal(content, { narrow = false } = {}) {
    closeModal();
    const root = $("#modal-root");
    root.innerHTML = `<div class="modal-backdrop"><div class="modal ${narrow ? "modal-narrow" : ""}">${content}</div></div>`;
    const backdrop = $(".modal-backdrop", root);
    backdrop.addEventListener("click", (event) => {
      if (event.target === backdrop) {
        closeModal();
      }
    });
    document.body.style.overflow = "hidden";
    escHandler = (event) => {
      if (event.key === "Escape") {
        closeModal();
      }
    };
    document.addEventListener("keydown", escHandler);
  }

  function closeModal() {
    const root = $("#modal-root");
    if (twitchImportTimer) {
      clearTimeout(twitchImportTimer);
      twitchImportTimer = null;
    }
    if (twitchImportState) {
      const importID = twitchImportState.id;
      twitchImportState = null;
      if (importID) {
        api(`/api/twitch/imports/${importID}`, { method: "DELETE", body: {} }).catch(() => {});
      }
    }
    if (!root.innerHTML) {
      return;
    }
    root.innerHTML = "";
    document.body.style.overflow = "";
    if (escHandler) {
      document.removeEventListener("keydown", escHandler);
      escHandler = null;
    }
    if (previewTimer) {
      clearTimeout(previewTimer);
      previewTimer = null;
    }
    if (routePreviewTimer) {
      clearTimeout(routePreviewTimer);
      routePreviewTimer = null;
    }
    previewSequence++;
    routePreviewSequence++;
    lastFocusedField = null;
  }

  function openWebhookModal(webhookId = "", { duplicate = false } = {}) {
    const source = webhookId ? webhookById(webhookId) : null;
    const existing = duplicate ? null : source;
    const prefill = source;
    const template = { ...(state.default_template || {}), ...(prefill?.template || {}) };
    const channels = state.config.channels || [];
    const previewOptions = channels.length
      ? channels.map((channel) => `<option value="${channel.id}">${escapeHTML(channel.display_name || channel.login)}</option>`).join("")
      : '<option value="">Sample stream</option>';
    const modalTitle = duplicate ? "Duplicate destination" : existing ? "Edit destination" : "New destination";
    const modalSubtitle = duplicate
      ? "Settings copied. Add a different Discord webhook URL and save."
      : existing
        ? "Change the label, webhook, or message style."
        : "Pick a Discord webhook and make the alert your own.";
    const labelValue = duplicate ? `${source?.label || "Destination"} copy` : (existing?.label || "");
    const urlValue = duplicate ? "" : (prefill?.url || "");
    const enabled = existing ? existing.enabled : true;
    const roleIDs = prefill?.role_ids || [];
    const pingStreamer = templatePingStreamer(template);
    const pingRoles = templatePingRoles(template);
    const updateSeconds = prefill?.update_seconds || 0;
    const cleanup = prefill?.cleanup || {};
    const cleanupMode = cleanup.mode || "off";
    const endedTemplate = { ...(state.default_ended_template || {}), ...(cleanup.template || {}) };

    openModal(`
      <div class="modal-head">
        <div><h2>${modalTitle}</h2><p>${modalSubtitle}</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="webhook-form" data-webhook-id="${existing?.id || ""}">
        <div class="modal-body">
          <div class="modal-split">
            <div class="form-grid">
              <label class="field">
                <span class="field-label">Label</span>
                <input id="wh-label" class="input" value="${escapeHTML(labelValue)}" placeholder="Community #live" autocomplete="off">
                <span class="field-hint">This is what you see when choosing where alerts go.</span>
              </label>
              <label class="field">
                <span class="field-label">Discord webhook URL <button type="button" class="placeholder-chip" data-action="toggle-url">Show</button></span>
                <input id="wh-url" class="input input-mono" type="password" value="${escapeHTML(urlValue)}" placeholder="https://discord.com/api/webhooks/..." autocomplete="off" spellcheck="false">
                <span class="field-hint">Server Settings → Integrations → Webhooks → New Webhook → Copy URL.</span>
              </label>
              <div class="form-grid two">
                <label class="field">
                  <span class="field-label">Webhook name</span>
                  <input id="wh-username" class="input" value="${escapeHTML(prefill?.username || "")}" placeholder="Twitch Live" autocomplete="off">
                  <span class="field-hint">Optional display name for this destination.</span>
                </label>
                <label class="field">
                  <span class="field-label">Avatar URL</span>
                  <input id="wh-avatar" class="input input-mono" value="${escapeHTML(prefill?.avatar_url || "")}" placeholder="https://example.com/avatar.png" autocomplete="off" spellcheck="false">
                  <span class="field-hint">Optional image shown beside the message.</span>
                </label>
              </div>
              <div class="form-grid two">
                <label class="field">
                  <span class="field-label">Thread ID</span>
                  <input id="wh-thread" class="input input-mono" value="${escapeHTML(prefill?.thread_id || "")}" placeholder="123456789012345678" autocomplete="off" spellcheck="false">
                  <span class="field-hint">Optional. Post into an existing Discord thread or forum post.</span>
                </label>
                <label class="field">
                  <span class="field-label">Link buttons</span>
                  <textarea id="wh-buttons" class="textarea" rows="4" placeholder="Watch on Twitch | {url}">${escapeHTML(buttonsText(prefill?.buttons))}</textarea>
                  <span class="field-hint">One per line: <code>Label | https://…</code>. Up to five.</span>
                </label>
              </div>
              <div class="switch-row">
                <div class="switch-copy"><strong>Destination active</strong><span>When paused, alerts are not sent here.</span></div>
                <label class="switch"><input id="wh-enabled" type="checkbox" ${enabled ? "checked" : ""}><span class="switch-track"></span></label>
              </div>

              <div class="settings-section">
                <div class="field-label" style="margin-bottom:7px">Mention controls</div>
                <label class="field">
                  <span class="field-label">Discord role IDs</span>
                  <input id="wh-roles" class="input input-mono" value="${escapeHTML(roleIDsText(roleIDs))}" placeholder="123456789012345678, 987654321098765432" autocomplete="off" spellcheck="false">
                  <span class="field-hint">Comma or space separated. Only these role IDs can be pinged by <code>{role_mentions}</code>.</span>
                </label>
                <div class="form-grid two">
                  ${toggleRow("tpl-ping-streamer", "Ping streamer", pingStreamer)}
                  ${toggleRow("tpl-ping-roles", "Ping configured roles", pingRoles)}
                </div>
                <div class="switch-row" style="padding:6px 0">
                  <div class="switch-copy"><strong>Allow @everyone / @here</strong><span>Off by default. Turning this on can ping the whole server.</span></div>
                  <label class="switch"><input id="tpl-allow-everyone" type="checkbox" ${template.allow_everyone ? "checked" : ""}><span class="switch-track"></span></label>
                </div>
                <div class="warning-box" style="margin-top:8px">${icon("bell")}<span>Role and user IDs are the only mentions the app can ping. <code>@everyone</code> and <code>@here</code> stay blocked unless this destination explicitly allows them.</span></div>
              </div>

              <div class="settings-section">
                <div class="field-label" style="margin-bottom:7px">Message content</div>
                <textarea id="tpl-content" class="textarea" maxlength="${MAX.content}" placeholder="🔴 **{streamer}** is live on Twitch!">${escapeHTML(template.content || "")}</textarea>
                <div class="field-label" style="margin-top:6px"><span>Discord message text</span><span class="counter" id="count-content">0/${MAX.content}</span></div>
                <div class="placeholder-row" style="margin-top:7px">
                  ${(state.placeholders || []).map((placeholder) => `<button type="button" class="placeholder-chip" data-action="insert-placeholder" data-token="${escapeHTML(placeholder.token)}" title="${escapeHTML(placeholder.description)}">${escapeHTML(placeholder.token)}</button>`).join("")}
                </div>
              </div>

              <div class="form-grid two">
                <label class="field">
                  <span class="field-label">Embed title</span>
                  <input id="tpl-title" class="input" maxlength="${MAX.title}" value="${escapeHTML(template.embed_title || "")}" placeholder="{title}">
                  <span class="field-hint"><span class="counter" id="count-title">0/${MAX.title}</span></span>
                </label>
                <label class="field">
                  <span class="field-label">Embed color</span>
                  <span style="display:flex;gap:8px;align-items:center">
                    <input id="tpl-color" type="color" value="${safeColor(template.color)}" style="width:44px;height:38px;padding:2px;border:1px solid var(--border);border-radius:9px;background:transparent">
                    <input id="tpl-color-text" class="input input-mono" value="${safeColor(template.color)}" style="flex:1">
                  </span>
                </label>
              </div>

              <label class="field">
                <span class="field-label">Embed description <span class="counter" id="count-description">0/${MAX.description}</span></span>
                <textarea id="tpl-description" class="textarea" maxlength="${MAX.description}">${escapeHTML(template.embed_description || "")}</textarea>
              </label>

              <label class="field">
                <span class="field-label">Footer <span class="counter" id="count-footer">0/${MAX.footer}</span></span>
                <input id="tpl-footer" class="input" maxlength="${MAX.footer}" value="${escapeHTML(template.footer || "")}">
              </label>

              <div class="settings-section">
                <div class="field-label" style="margin-bottom:4px">Embed details</div>
                <div class="form-grid two">
                  ${toggleRow("tpl-author", "Author", template.show_author)}
                  ${toggleRow("tpl-image", "Preview image", template.show_image)}
                  ${toggleRow("tpl-game", "Game field", template.show_game)}
                  ${toggleRow("tpl-viewers", "Viewers field", template.show_viewers)}
                  ${toggleRow("tpl-live", "Live time field", template.show_live_since)}
                </div>
              </div>

              <div class="settings-section">
                <div class="field-label" style="margin-bottom:7px">Custom fields and order</div>
                <div class="form-grid two">
                  <label class="field">
                    <span class="field-label">Custom embed fields</span>
                    <textarea id="tpl-custom-fields" class="textarea" rows="3" placeholder="Schedule | Every Saturday | inline">${escapeHTML(customFieldsText(template.custom_fields))}</textarea>
                    <span class="field-hint">One per line: <code>Name | Value | inline</code>. Placeholders work in both columns.</span>
                  </label>
                  <label class="field">
                    <span class="field-label">Field order</span>
                    <input id="tpl-field-order" class="input input-mono" value="${escapeHTML(fieldOrderText(template.field_order) || defaultFieldOrder(template))}" placeholder="game, viewers, live, custom:0">
                    <span class="field-hint">Comma separated: <code>game</code>, <code>viewers</code>, <code>live</code>, <code>custom:0</code>.</span>
                  </label>
                </div>
              </div>

              <div class="settings-section">
                <div class="field-label" style="margin-bottom:7px">Stream lifecycle</div>
                <div class="form-grid two">
                  <label class="field">
                    <span class="field-label">Update the live message</span>
                    <select id="wh-update-seconds" class="select">
                      ${updateIntervalOptions(updateSeconds)}
                    </select>
                    <span class="field-hint">Edits the same Discord message. Off leaves it untouched.</span>
                  </label>
                  <label class="field">
                    <span class="field-label">When the stream ends</span>
                    <select id="wh-cleanup-mode" class="select">
                      <option value="off" ${cleanupMode === "off" ? "selected" : ""}>Leave the live message</option>
                      <option value="delete" ${cleanupMode === "delete" ? "selected" : ""}>Delete the live message</option>
                      <option value="edit" ${cleanupMode === "edit" ? "selected" : ""}>Edit it into an ended message</option>
                      <option value="post" ${cleanupMode === "post" ? "selected" : ""}>Post a separate ended message</option>
                    </select>
                    <span class="field-hint">Uses the original message and webhook token, even if the URL changes later.</span>
                  </label>
                </div>
                <div id="ended-template-fields" class="settings-section">
                  <div class="field-label" style="margin-bottom:7px">Stream-ended message</div>
                  <label class="field">
                    <span class="field-label">Message content</span>
                    <textarea id="end-content" class="textarea" maxlength="${MAX.content}">${escapeHTML(endedTemplate.content || "")}</textarea>
                  </label>
                  <div class="form-grid two">
                    <label class="field">
                      <span class="field-label">Embed title</span>
                      <input id="end-title" class="input" maxlength="${MAX.title}" value="${escapeHTML(endedTemplate.embed_title || "")}">
                    </label>
                    <label class="field">
                      <span class="field-label">Embed color</span>
                      <span style="display:flex;gap:8px;align-items:center">
                        <input id="end-color" type="color" value="${safeColor(endedTemplate.color)}" style="width:44px;height:38px;padding:2px;border:1px solid var(--border);border-radius:9px;background:transparent">
                        <input id="end-color-text" class="input input-mono" value="${safeColor(endedTemplate.color)}" style="flex:1">
                      </span>
                    </label>
                  </div>
                  <label class="field">
                    <span class="field-label">Embed description</span>
                    <textarea id="end-description" class="textarea" maxlength="${MAX.description}">${escapeHTML(endedTemplate.embed_description || "")}</textarea>
                    <span class="field-hint">Try <code>{vod}</code> for the latest VOD link, or <code>{title}</code> and <code>{game}</code>.</span>
                  </label>
                  <label class="field">
                    <span class="field-label">Footer</span>
                    <input id="end-footer" class="input" maxlength="${MAX.footer}" value="${escapeHTML(endedTemplate.footer || "")}">
                  </label>
                  <div class="switch-row" style="padding:6px 0">
                    <div class="switch-copy"><strong>Preview the ended message</strong><span>Shows what Discord will display after the stream ends.</span></div>
                    <label class="switch"><input id="preview-ended" type="checkbox"><span class="switch-track"></span></label>
                  </div>
                </div>
              </div>
            </div>

            <div class="preview-panel">
              <h3>Live Discord preview</h3>
              <select id="preview-channel" class="select preview-select">${previewOptions}</select>
              <div id="discord-preview" class="discord-preview"></div>
              <div id="preview-warning"></div>
            </div>
          </div>
        </div>
        <div class="modal-foot">
          <button type="button" class="btn btn-ghost" data-action="universal-close">Cancel</button>
          <div class="modal-foot-right">
            <button type="button" class="btn btn-ghost" data-action="reset-template">${icon("refresh")} Reset style</button>
            <button type="button" class="btn btn-ghost" data-action="export-template">${icon("download")} Export</button>
            <button type="button" class="btn btn-ghost" data-action="import-template">${icon("upload")} Import</button>
            <input id="template-import" type="file" accept="application/json,.json" style="display:none">
            ${existing ? `<button type="button" class="btn btn-ghost" data-action="duplicate-webhook" data-webhook="${existing.id}">${icon("copy")} Duplicate</button>` : ""}
            ${existing ? `<button type="button" class="btn btn-ghost" data-action="test-webhook" data-webhook="${existing.id}">${icon("send")} Send test</button>` : ""}
            <button type="submit" class="btn btn-primary">${icon("check")} ${existing ? "Save destination" : "Add destination"}</button>
          </div>
        </div>
      </form>
    `);

    bindWebhookEditor();
  }

  function toggleRow(id, label, checked) {
    return `<div class="switch-row" style="padding:6px 0">
      <div class="switch-copy"><strong>${escapeHTML(label)}</strong></div>
      <label class="switch"><input id="${id}" type="checkbox" ${checked ? "checked" : ""}><span class="switch-track"></span></label>
    </div>`;
  }

  function bindWebhookEditor() {
    const panel = $(".modal");
    if (!panel) {
      return;
    }
    lastFocusedField = $("#tpl-content");
    panel.addEventListener("focusin", (event) => {
      if (event.target.matches("input, textarea, select")) {
        lastFocusedField = event.target;
      }
    });
    $$("input, textarea, select", panel).forEach((field) => {
      field.addEventListener("input", () => {
        updateCounters();
        schedulePreview();
      });
      field.addEventListener("change", () => {
        updateCounters();
        schedulePreview();
      });
    });
    const colorPairs = [
      ["#tpl-color-text", "#tpl-color"],
      ["#end-color-text", "#end-color"],
    ];
    colorPairs.forEach(([textSelector, pickerSelector]) => {
      const colorText = $(textSelector);
      const colorPicker = $(pickerSelector);
      if (!colorText || !colorPicker) {
        return;
      }
      colorText.addEventListener("input", () => {
        if (/^#[0-9a-fA-F]{6}$/.test(colorText.value)) {
          colorPicker.value = colorText.value;
        }
      });
      colorPicker.addEventListener("input", () => {
        colorText.value = colorPicker.value;
      });
    });
    const cleanupMode = $("#wh-cleanup-mode");
    if (cleanupMode) {
      cleanupMode.addEventListener("change", syncEndedFields);
    }
    const urlField = $("#wh-url");
    if (urlField) {
      urlField.addEventListener("input", scheduleWebhookInspect);
      urlField.addEventListener("blur", () => {
        clearTimeout(webhookInspectTimer);
        webhookInspectTimer = null;
        inspectWebhookURL();
      });
    }
    syncEndedFields();
    updateCounters();
    schedulePreview();
  }

  function scheduleWebhookInspect() {
    clearTimeout(webhookInspectTimer);
    webhookInspectTimer = setTimeout(() => {
      webhookInspectTimer = null;
      inspectWebhookURL();
    }, 650);
  }

  async function inspectWebhookURL() {
    const urlField = $("#wh-url");
    const labelField = $("#wh-label");
    if (!urlField || !labelField) {
      return;
    }
    const raw = urlField.value.trim();
    if (!raw || labelField.value.trim()) {
      return;
    }
    if (!/^https:\/\/(?:[a-z0-9-]+\.)?discord(?:app)?\.com\/api\/webhooks\/\d+\//i.test(raw)) {
      return;
    }
    const sequence = ++webhookInspectSequence;
    try {
      const data = await api("/api/webhooks/inspect", { method: "POST", body: { url: raw } });
      if (sequence !== webhookInspectSequence) {
        return;
      }
      const currentLabel = $("#wh-label");
      if (!currentLabel || currentLabel.value.trim()) {
        return;
      }
      const suggestion = String(data?.name || "").trim();
      if (!suggestion) {
        return;
      }
      currentLabel.value = truncate(suggestion, 80);
      currentLabel.dispatchEvent(new Event("input", { bubbles: true }));
      toast("Label suggested from Discord. You can edit it.", "info");
    } catch {
      // The inspect endpoint is optional; failing silently keeps manual entry intact.
    }
  }

  function openRouteTemplateModal(channelId, webhookId) {
    const channel = channelById(channelId);
    const webhook = webhookById(webhookId);
    if (!channel || !webhook) {
      return;
    }
    const destination = (channel.destinations || []).find((item) => item.webhook_id === webhookId);
    const custom = Boolean(destination?.template);
    const template = { ...(state.default_template || {}), ...(custom ? destination.template : webhook.template) };
    const previewOptions = (state.config.channels || [])
      .map((item) => `<option value="${item.id}" ${item.id === channel.id ? "selected" : ""}>${escapeHTML(item.display_name || item.login)}</option>`)
      .join("") || '<option value="">Sample stream</option>';

    openModal(`
      <div class="modal-head">
        <div><h2>Route template</h2><p>${escapeHTML(channel.display_name || channel.login)} → ${escapeHTML(webhook.label)}. This only changes this route.</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="route-template-form" data-channel-id="${channel.id}" data-webhook-id="${webhook.id}">
        <div class="modal-body">
          <div class="modal-split">
            <div class="form-grid">
              <div class="warning-box">${icon("bell")}<span>${custom ? "This route has its own template. Clearing it falls back to the destination template." : "Saving here creates an override for this route only. The destination template stays unchanged."}</span></div>
              <label class="field">
                <span class="field-label">Message content</span>
                <textarea id="route-content" class="textarea" maxlength="${MAX.content}">${escapeHTML(template.content || "")}</textarea>
              </label>
              <div class="form-grid two">
                <label class="field">
                  <span class="field-label">Embed title</span>
                  <input id="route-title" class="input" maxlength="${MAX.title}" value="${escapeHTML(template.embed_title || "")}">
                </label>
                <label class="field">
                  <span class="field-label">Embed color</span>
                  <span style="display:flex;gap:8px;align-items:center">
                    <input id="route-color" type="color" value="${safeColor(template.color)}" style="width:44px;height:38px;padding:2px;border:1px solid var(--border);border-radius:9px;background:transparent">
                    <input id="route-color-text" class="input input-mono" value="${safeColor(template.color)}" style="flex:1">
                  </span>
                </label>
              </div>
              <label class="field">
                <span class="field-label">Embed description</span>
                <textarea id="route-description" class="textarea" maxlength="${MAX.description}">${escapeHTML(template.embed_description || "")}</textarea>
              </label>
              <label class="field">
                <span class="field-label">Footer</span>
                <input id="route-footer" class="input" maxlength="${MAX.footer}" value="${escapeHTML(template.footer || "")}">
              </label>
              <div class="settings-section">
                <div class="field-label" style="margin-bottom:4px">Embed details</div>
                <div class="form-grid two">
                  ${toggleRow("route-author", "Author", template.show_author)}
                  ${toggleRow("route-image", "Preview image", template.show_image)}
                  ${toggleRow("route-game", "Game field", template.show_game)}
                  ${toggleRow("route-viewers", "Viewers field", template.show_viewers)}
                  ${toggleRow("route-live", "Live time field", template.show_live_since)}
                </div>
              </div>
              <div class="form-grid two">
                <label class="field">
                  <span class="field-label">Custom embed fields</span>
                  <textarea id="route-custom-fields" class="textarea" rows="3" placeholder="Schedule | Every Saturday | inline">${escapeHTML(customFieldsText(template.custom_fields))}</textarea>
                  <span class="field-hint">One per line: <code>Name | Value | inline</code>.</span>
                </label>
                <label class="field">
                  <span class="field-label">Field order</span>
                  <input id="route-field-order" class="input input-mono" value="${escapeHTML(fieldOrderText(template.field_order) || defaultFieldOrder(template))}" placeholder="game, viewers, live, custom:0">
                  <span class="field-hint">Comma separated built-ins and <code>custom:N</code>.</span>
                </label>
              </div>
              <div class="placeholder-row">
                ${(state.placeholders || []).map((placeholder) => `<button type="button" class="placeholder-chip" data-action="insert-route-placeholder" data-token="${escapeHTML(placeholder.token)}" title="${escapeHTML(placeholder.description)}">${escapeHTML(placeholder.token)}</button>`).join("")}
              </div>
            </div>
            <div class="preview-panel">
              <h3>Live route preview</h3>
              <select id="route-preview-channel" class="select preview-select">${previewOptions}</select>
              <div id="route-discord-preview" class="discord-preview"></div>
              <div id="route-preview-warning"></div>
            </div>
          </div>
        </div>
        <div class="modal-foot">
          <button type="button" class="btn btn-ghost" data-action="universal-close">Cancel</button>
          <div class="modal-foot-right">
            ${custom ? `<button type="button" class="btn btn-ghost" data-action="clear-route-template" data-channel="${channel.id}" data-webhook="${webhook.id}">${icon("refresh")} Use destination template</button>` : ""}
            <button type="submit" class="btn btn-primary">${icon("check")} Save route template</button>
          </div>
        </div>
      </form>
    `);

    const panel = $(".modal");
    if (!panel) {
      return;
    }
    panel.addEventListener("focusin", (event) => {
      if (event.target.matches("input, textarea, select")) {
        routeLastFocusedField = event.target;
      }
    });
    $$("input, textarea, select", panel).forEach((field) => {
      field.addEventListener("input", scheduleRoutePreview);
      field.addEventListener("change", scheduleRoutePreview);
    });
    const colorPairs = [
      ["#route-color-text", "#route-color"],
    ];
    colorPairs.forEach(([textSelector, pickerSelector]) => {
      const colorText = $(textSelector);
      const colorPicker = $(pickerSelector);
      if (!colorText || !colorPicker) {
        return;
      }
      colorText.addEventListener("input", () => {
        if (/^#[0-9a-fA-F]{6}$/.test(colorText.value)) {
          colorPicker.value = colorText.value;
        }
      });
      colorPicker.addEventListener("input", () => {
        colorText.value = colorPicker.value;
      });
    });
    routeLastFocusedField = $("#route-content");
    scheduleRoutePreview();
  }

  function readRouteTemplate() {
    return {
      content: $("#route-content")?.value || "",
      embed_title: $("#route-title")?.value || "",
      embed_description: $("#route-description")?.value || "",
      footer: $("#route-footer")?.value || "",
      color: safeColor($("#route-color-text")?.value || $("#route-color")?.value),
      show_author: Boolean($("#route-author")?.checked),
      show_image: Boolean($("#route-image")?.checked),
      show_game: Boolean($("#route-game")?.checked),
      show_viewers: Boolean($("#route-viewers")?.checked),
      show_live_since: Boolean($("#route-live")?.checked),
      custom_fields: parseCustomFields($("#route-custom-fields")?.value || ""),
      field_order: parseFieldOrder($("#route-field-order")?.value || ""),
    };
  }

  function scheduleRoutePreview() {
    if (routePreviewTimer) {
      clearTimeout(routePreviewTimer);
    }
    routePreviewTimer = setTimeout(updateRoutePreview, 180);
  }

  async function updateRoutePreview() {
    const preview = $("#route-discord-preview");
    if (!preview) {
      return;
    }
    const sequence = ++routePreviewSequence;
    const form = $("#route-template-form");
    const channelId = $("#route-preview-channel")?.value || "";
    const webhook = webhookById(form?.dataset.webhookId || "");
    const template = readRouteTemplate();
    const destinationTemplate = webhook?.template || {};
    template.ping_streamer = destinationTemplate.ping_streamer;
    template.ping_roles = destinationTemplate.ping_roles;
    template.allow_everyone = destinationTemplate.allow_everyone;
    try {
      const data = await api("/api/preview", {
        method: "POST",
        body: {
          template,
          channel_id: channelId,
          webhook_label: webhook?.label || "Discord",
          role_ids: webhook?.role_ids || [],
          username: webhook?.username || "",
          avatar_url: webhook?.avatar_url || "",
          thread_id: webhook?.thread_id || "",
          buttons: webhook?.buttons || [],
        },
      });
      if (sequence !== routePreviewSequence) {
        return;
      }
      preview.innerHTML = renderDiscordPreview(data.payload, Boolean($("#route-image")?.checked));
      const unknown = data.unknown_placeholders || [];
      const currentTemplate = readRouteTemplate();
      const empty = !currentTemplate.content.trim() && !currentTemplate.embed_title.trim() && !currentTemplate.embed_description.trim() && !(currentTemplate.custom_fields || []).length;
      const messages = [];
      if (unknown.length) {
        messages.push(`Unknown placeholder${unknown.length === 1 ? "" : "s"}: ${unknown.join(", ")}`);
      }
      if (empty) {
        messages.push("Nothing will be posted. Add message content, embed text, or a custom field.");
      }
      $("#route-preview-warning").innerHTML = messages.map((message) => `<div class="warning-box" style="margin-top:10px">${icon("bell")} <span>${escapeHTML(message)}</span></div>`).join("");
    } catch (error) {
      if (sequence !== routePreviewSequence) {
        return;
      }
      preview.innerHTML = `<div class="empty-state" style="min-height:140px">${escapeHTML(error.message)}</div>`;
    }
  }

  async function clearRouteTemplate(channelId, webhookId) {
    if (!window.confirm("Use the destination template for this route? The custom override will be removed.")) {
      return;
    }
    try {
      state = await api(`/api/channels/${channelId}/destinations/${webhookId}`, {
        method: "PUT",
        body: { clear: true },
      });
      closeModal();
      render();
      toast("Route template override removed.", "info");
    } catch (error) {
      toast(error.message, "error");
    }
  }

  function readTemplate() {
    return {
      content: $("#tpl-content")?.value || "",
      embed_title: $("#tpl-title")?.value || "",
      embed_description: $("#tpl-description")?.value || "",
      footer: $("#tpl-footer")?.value || "",
      color: safeColor($("#tpl-color-text")?.value || $("#tpl-color")?.value),
      show_author: Boolean($("#tpl-author")?.checked),
      show_image: Boolean($("#tpl-image")?.checked),
      show_game: Boolean($("#tpl-game")?.checked),
      show_viewers: Boolean($("#tpl-viewers")?.checked),
      show_live_since: Boolean($("#tpl-live")?.checked),
      ping_streamer: Boolean($("#tpl-ping-streamer")?.checked),
      ping_roles: Boolean($("#tpl-ping-roles")?.checked),
      allow_everyone: Boolean($("#tpl-allow-everyone")?.checked),
      custom_fields: parseCustomFields($("#tpl-custom-fields")?.value || ""),
      field_order: parseFieldOrder($("#tpl-field-order")?.value || ""),
    };
  }

  function readEndedTemplate() {
    return {
      content: $("#end-content")?.value || "",
      embed_title: $("#end-title")?.value || "",
      embed_description: $("#end-description")?.value || "",
      footer: $("#end-footer")?.value || "",
      color: safeColor($("#end-color-text")?.value || $("#end-color")?.value),
      show_author: true,
      show_image: false,
      show_game: false,
      show_viewers: false,
      show_live_since: true,
      ping_streamer: false,
      ping_roles: false,
      allow_everyone: false,
    };
  }

  function endedTemplateEnabled() {
    const mode = $("#wh-cleanup-mode")?.value || "off";
    return mode === "edit" || mode === "post";
  }

  function syncEndedFields() {
    const fields = $("#ended-template-fields");
    if (fields) {
      fields.style.display = endedTemplateEnabled() ? "" : "none";
    }
  }

  function fillTemplate(template) {
    const fields = {
      "tpl-content": template.content || "",
      "tpl-title": template.embed_title || "",
      "tpl-description": template.embed_description || "",
      "tpl-footer": template.footer || "",
    };
    Object.entries(fields).forEach(([id, value]) => {
      const field = $(`#${id}`);
      if (field) {
        field.value = value;
      }
    });
    const color = safeColor(template.color);
    if ($("#tpl-color")) {
      $("#tpl-color").value = color;
    }
    if ($("#tpl-color-text")) {
      $("#tpl-color-text").value = color;
    }
    if ($("#tpl-custom-fields")) {
      $("#tpl-custom-fields").value = customFieldsText(template.custom_fields);
    }
    if ($("#tpl-field-order")) {
      $("#tpl-field-order").value = fieldOrderText(template.field_order) || defaultFieldOrder(template);
    }
    const toggles = {
      "tpl-author": template.show_author,
      "tpl-image": template.show_image,
      "tpl-game": template.show_game,
      "tpl-viewers": template.show_viewers,
      "tpl-live": template.show_live_since,
      "tpl-ping-streamer": templatePingStreamer(template),
      "tpl-ping-roles": templatePingRoles(template),
      "tpl-allow-everyone": template.allow_everyone,
    };
    Object.entries(toggles).forEach(([id, checked]) => {
      const field = $(`#${id}`);
      if (field) {
        field.checked = Boolean(checked);
      }
    });
    updateCounters();
    schedulePreview();
  }

  function updateCounters() {
    const mappings = [
      ["#tpl-content", "#count-content", MAX.content],
      ["#tpl-title", "#count-title", MAX.title],
      ["#tpl-description", "#count-description", MAX.description],
      ["#tpl-footer", "#count-footer", MAX.footer],
    ];
    mappings.forEach(([fieldSelector, counterSelector, max]) => {
      const field = $(fieldSelector);
      const counter = $(counterSelector);
      if (!field || !counter) {
        return;
      }
      const length = field.value.length;
      counter.textContent = `${length}/${max}`;
      counter.classList.toggle("is-over", length > max);
    });
  }

  function schedulePreview() {
    if (previewTimer) {
      clearTimeout(previewTimer);
    }
    previewTimer = setTimeout(updatePreview, 180);
  }

  async function updatePreview() {
    const preview = $("#discord-preview");
    if (!preview) {
      return;
    }
    const sequence = ++previewSequence;
    const endedPreview = Boolean($("#preview-ended")?.checked) && endedTemplateEnabled();
    const template = endedPreview ? readEndedTemplate() : readTemplate();
    const channelId = $("#preview-channel")?.value || "";
    const webhookLabel = $("#wh-label")?.value || "Discord";
    const roleIDs = parseRoleIDs($("#wh-roles")?.value || "");
    try {
      const data = await api("/api/preview", {
        method: "POST",
        body: {
          template,
          channel_id: channelId,
          webhook_label: webhookLabel,
          role_ids: roleIDs,
          username: $("#wh-username")?.value || "",
          avatar_url: $("#wh-avatar")?.value || "",
          thread_id: $("#wh-thread")?.value || "",
          buttons: parseButtons($("#wh-buttons")?.value || ""),
          ended: endedPreview,
        },
      });
      if (sequence !== previewSequence) {
        return;
      }
      preview.innerHTML = renderDiscordPreview(data.payload, template.show_image);
      const warning = $("#preview-warning");
      const unknown = data.unknown_placeholders || [];
      const empty = !template.content.trim() && !template.embed_title.trim() && !template.embed_description.trim();
      const previewChannel = channelById(channelId);
      const allText = [template.content, template.embed_title, template.embed_description, template.footer].join("\n");
      const rendered = data.payload || {};
      const renderedText = [
        rendered.content,
        rendered.embeds?.[0]?.title,
        rendered.embeds?.[0]?.description,
        rendered.embeds?.[0]?.footer?.text,
      ].filter(Boolean).join("\n");
      const usesDiscordMention = allText.includes("{discord_mention}");
      const usesRoleMentions = allText.includes("{role_mentions}");
      const missingDiscordID = usesDiscordMention && template.ping_streamer && channelId && previewChannel && !previewChannel.discord_id;
      const missingRoles = usesRoleMentions && template.ping_roles && !roleIDs.length;
      const everyoneTyped = /@(everyone|here)\b/i.test(renderedText);
      const messages = [];
      if (unknown.length) {
        messages.push(`Unknown placeholder${unknown.length === 1 ? "" : "s"}: ${unknown.join(", ")}`);
      }
      if (empty) {
        messages.push("Nothing will be posted. Add message content or embed text.");
      }
      if (missingDiscordID) {
        messages.push("This streamer has no Discord user ID yet. Add one under Streamers, or {discord_mention} will render as plain @text.");
      }
      if (missingRoles) {
        messages.push("Add at least one role ID to use {role_mentions}.");
      }
      if (roleIDs.length && !template.ping_roles) {
        messages.push("Role pings are turned off, so {role_mentions} renders empty.");
      }
      if (everyoneTyped && !template.allow_everyone) {
        messages.push("@everyone / @here will not ping. Enable it explicitly in Mention controls if you really want it.");
      }
      if (template.allow_everyone) {
        messages.push("This destination is allowed to ping @everyone / @here. Use with care.");
      }
      const allowed = data.payload?.allowed_mentions || { parse: [], roles: [], users: [] };
      const pings = [];
      const seenPings = new Set();
      for (const match of renderedText.matchAll(/<@!?(\d{17,20})>/g)) {
        if ((allowed.users || []).includes(match[1]) && !seenPings.has(`user:${match[1]}`)) {
          seenPings.add(`user:${match[1]}`);
          pings.push(`@user ${match[1]}`);
        }
      }
      for (const match of renderedText.matchAll(/<@&(\d{17,20})>/g)) {
        if ((allowed.roles || []).includes(match[1]) && !seenPings.has(`role:${match[1]}`)) {
          seenPings.add(`role:${match[1]}`);
          pings.push(`@role ${match[1]}`);
        }
      }
      if (everyoneTyped && (allowed.parse || []).includes("everyone")) {
        pings.push("@everyone / @here");
      }
      const warningsHTML = messages.map((message) => `<div class="warning-box" style="margin-top:10px">${icon("bell")} <span>${escapeHTML(message)}</span></div>`).join("");
      const pingsHTML = pings.length
        ? `<div class="ping-summary">Will ping: ${pings.map((ping) => `<span class="ping-chip">${escapeHTML(ping)}</span>`).join("")}</div>`
        : "";
      warning.innerHTML = warningsHTML + pingsHTML;
    } catch (error) {
      if (sequence !== previewSequence) {
        return;
      }
      preview.innerHTML = `<div class="empty-state" style="min-height:140px">${escapeHTML(error.message)}</div>`;
    }
  }

  function renderDiscordPreview(payload, showImage = true) {
    const embed = payload?.embeds?.[0];
    const content = payload?.content || "";
    const color = hexFromInt(embed?.color);
    const author = embed?.author || {};
    const image = safeURL(embed?.image?.url || "");
    const authorIcon = safeURL(author.icon_url || "");
    const fields = (embed?.fields || []).map((field) => `
      <div class="discord-field">
        <div class="discord-field-name">${escapeHTML(field.name)}</div>
        <div class="discord-field-value">${discordMarkdown(field.value)}</div>
      </div>`).join("");
    const title = embed?.title ? (embed.url ? `<a href="${escapeHTML(safeURL(embed.url))}" target="_blank" rel="noopener">${escapeHTML(embed.title)}</a>` : escapeHTML(embed.title)) : "";
    const imageHTML = image ? `<img class="discord-embed-image" src="${escapeHTML(image)}" alt="">` : "";
    const footer = embed?.footer?.text ? `<div class="discord-footer">${escapeHTML(embed.footer.text)}</div>` : "";
    const timestamp = new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
    const authorName = payload?.username || "Twitch Live";
    const authorAvatar = safeURL(payload?.avatar_url || "") || "/static/icon.png";
    const buttons = (payload?.components?.[0]?.components || [])
      .map((button) => `<a class="discord-button" href="${escapeHTML(safeURL(button.url))}" target="_blank" rel="noopener">${escapeHTML(button.label)}</a>`)
      .join("");
    return `
      <div class="discord-message">
        <div class="discord-avatar"><img src="${escapeHTML(authorAvatar)}" alt=""></div>
        <div class="discord-message-body">
          <div class="discord-message-head"><span class="discord-author">${escapeHTML(authorName)}</span><span class="discord-bot-tag">App</span><span class="discord-timestamp">Today at ${escapeHTML(timestamp)}</span></div>
          ${content ? `<div class="discord-content">${discordMarkdown(content)}</div>` : ""}
          ${embed ? `<div class="discord-embed" style="border-left-color:${color}">
            ${author.name ? `<div class="discord-embed-author">${authorIcon ? `<img src="${escapeHTML(authorIcon)}" alt="">` : ""}${escapeHTML(author.name)}</div>` : ""}
            ${title ? `<div class="discord-embed-title">${title}</div>` : ""}
            ${embed.description ? `<div class="discord-embed-desc">${discordMarkdown(embed.description)}</div>` : ""}
            ${fields ? `<div class="discord-fields">${fields}</div>` : ""}
            ${imageHTML || (showImage ? `<div class="discord-empty-image">Twitch preview appears here</div>` : "")}
            ${footer}
          </div>` : ""}
          ${buttons ? `<div class="discord-buttons">${buttons}</div>` : ""}
        </div>
      </div>`;
  }

  function formatDiscordTimestamp(unix, style) {
    const date = new Date(Number(unix) * 1000);
    const key = String(style);
    if (key.toLowerCase() === "r") {
      return formatRelative(date.toISOString());
    }
    const options = {
      f: { dateStyle: "medium", timeStyle: "short" },
      F: { dateStyle: "long", timeStyle: "short" },
      d: { dateStyle: "short" },
      D: { dateStyle: "long" },
      t: { timeStyle: "short" },
      T: { timeStyle: "medium" },
    }[key] || { dateStyle: "medium", timeStyle: "short" };
    return new Intl.DateTimeFormat(undefined, options).format(date);
  }

  function discordMarkdown(value) {
    let text = escapeHTML(value || "");
    text = text.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
    text = text.replace(/\*([^*]+)\*/g, "<em>$1</em>");
    text = text.replace(/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');
    text = text.replace(/&lt;t:(\d+):([RrFfDdTt])&gt;/g, (_, unix, style) => formatDiscordTimestamp(unix, style));
    text = text.replace(/&lt;@&(\d+)&gt;/g, '<span class="discord-mention">@role</span>');
    text = text.replace(/&lt;@!?(\d+)&gt;/g, '<span class="discord-mention">@user</span>');
    text = text.replace(/(^|[\s(])@(everyone|here)\b/gi, '$1<span class="discord-mention">@$2</span>');
    return text.replace(/\n/g, "<br>");
  }

  function openFiltersModal(channelId) {
    const channel = channelById(channelId);
    if (!channel) {
      return;
    }
    const f = channel.filters || {};
    const days = DAY_NAMES.map((label, index) => `
      <label class="destination-check">
        <input type="checkbox" name="filter-day" value="${index}" ${(f.allowed_days || []).includes(index) ? "checked" : ""}>
        <span>${label}</span>
      </label>`).join("");
    openModal(`
      <div class="modal-head">
        <div><h2>Alert filters</h2><p>Decide when ${escapeHTML(channel.display_name || channel.login)} should post. Live status still updates when an alert is skipped.</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="filters-form" data-channel-id="${channel.id}">
        <div class="modal-body">
          <div class="form-grid two">
            <label class="field">
              <span class="field-label">Only these games/categories</span>
              <input id="filter-include-games" class="input" value="${escapeHTML((f.include_games || []).join(", "))}" placeholder="Just Chatting, Music" autocomplete="off">
              <span class="field-hint">Comma separated. Case-insensitive contains match. Leave blank for all games.</span>
            </label>
            <label class="field">
              <span class="field-label">Never these games/categories</span>
              <input id="filter-exclude-games" class="input" value="${escapeHTML((f.exclude_games || []).join(", "))}" placeholder="Just Chatting" autocomplete="off">
              <span class="field-hint">Exclude always wins over include.</span>
            </label>
            <label class="field">
              <span class="field-label">Title must contain</span>
              <input id="filter-include-titles" class="input" value="${escapeHTML((f.include_titles || []).join(", "))}" placeholder="speedrun, ranked" autocomplete="off">
              <span class="field-hint">Comma separated keywords.</span>
            </label>
            <label class="field">
              <span class="field-label">Title must not contain</span>
              <input id="filter-exclude-titles" class="input" value="${escapeHTML((f.exclude_titles || []).join(", "))}" placeholder="spoilers, rerun" autocomplete="off">
              <span class="field-hint">Exclude always wins over include.</span>
            </label>
          </div>

          <div class="settings-section">
            <div class="field-label" style="margin-bottom:7px">Viewers and timing</div>
            <div class="form-grid two">
              <label class="field">
                <span class="field-label">Minimum viewers</span>
                <input id="filter-min-viewers" class="input" type="number" min="0" step="1" value="${Number(f.min_viewers) || 0}">
                <span class="field-hint">0 means no minimum.</span>
              </label>
              <label class="field">
                <span class="field-label">Alert delay (minutes)</span>
                <input id="filter-delay" class="input" type="number" min="0" max="10080" step="1" value="${Number(f.delay_minutes) || 0}">
                <span class="field-hint">Wait until the stream has been live this long.</span>
              </label>
              <label class="field">
                <span class="field-label">Cooldown between new streams (minutes)</span>
                <input id="filter-cooldown" class="input" type="number" min="0" max="10080" step="1" value="${Number(f.cooldown_minutes) || 0}">
                <span class="field-hint">0 means no cooldown.</span>
              </label>
              <div class="field">
                <span class="field-label">Only the first stream of the day</span>
                <div class="switch-row" style="padding:6px 0">
                  <div class="switch-copy"><span>Post once per local calendar day.</span></div>
                  <label class="switch"><input id="filter-first-stream" type="checkbox" ${f.first_stream_of_day ? "checked" : ""}><span class="switch-track"></span></label>
                </div>
              </div>
            </div>
          </div>

          <div class="settings-section">
            <div class="field-label" style="margin-bottom:7px">Quiet hours and days</div>
            <div class="form-grid two">
              <label class="field">
                <span class="field-label">Quiet from</span>
                <input id="filter-quiet-start" class="input" type="time" value="${escapeHTML(f.quiet_start || "")}">
                <span class="field-hint">Local timezone. Leave both blank to disable.</span>
              </label>
              <label class="field">
                <span class="field-label">Quiet until</span>
                <input id="filter-quiet-end" class="input" type="time" value="${escapeHTML(f.quiet_end || "")}">
                <span class="field-hint">Windows can cross midnight.</span>
              </label>
            </div>
            <div class="field-label" style="margin:8px 0 7px">Allowed days</div>
            <div class="destination-checks">${days}</div>
            <p class="field-hint" style="margin:9px 0 0">Leave every day unchecked to allow all days.</p>
          </div>

          <div class="warning-box" style="margin-top:12px">${icon("bell")}<span>Filters only affect posting. The streamer still shows as live, and dedupe still applies per destination.</span></div>
        </div>
        <div class="modal-foot">
          <button type="button" class="btn btn-ghost" data-action="universal-close">Cancel</button>
          <button type="submit" class="btn btn-primary">${icon("check")} Save filters</button>
        </div>
      </form>
    `);
  }

  function readFilters() {
    return {
      include_games: listFrom($("#filter-include-games")?.value),
      exclude_games: listFrom($("#filter-exclude-games")?.value),
      include_titles: listFrom($("#filter-include-titles")?.value),
      exclude_titles: listFrom($("#filter-exclude-titles")?.value),
      min_viewers: Math.max(0, Math.round(Number($("#filter-min-viewers")?.value) || 0)),
      delay_minutes: Math.max(0, Math.round(Number($("#filter-delay")?.value) || 0)),
      cooldown_minutes: Math.max(0, Math.round(Number($("#filter-cooldown")?.value) || 0)),
      first_stream_of_day: Boolean($("#filter-first-stream")?.checked),
      quiet_start: $("#filter-quiet-start")?.value || "",
      quiet_end: $("#filter-quiet-end")?.value || "",
      allowed_days: $$('input[name="filter-day"]:checked').map((input) => Number(input.value)),
    };
  }

  function openDiscordIDModal(channelId) {
    const channel = channelById(channelId);
    if (!channel) {
      return;
    }
    openModal(`
      <div class="modal-head">
        <div><h2>Discord ping</h2><p>Optional. Associate ${escapeHTML(channel.display_name || channel.login)} with a Discord user ID.</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="discord-id-form" data-channel-id="${channel.id}">
        <div class="modal-body">
          <label class="field">
            <span class="field-label">Discord user ID</span>
            <input id="discord-user-id" class="input input-mono" value="${escapeHTML(channel.discord_id || "")}" placeholder="123456789012345678" autocomplete="off" spellcheck="false">
            <span class="field-hint">17–20 digit Discord user ID. A profile URL or &lt;@mention&gt; also works. Leave blank to remove the ping.</span>
          </label>
          <div class="warning-box" style="margin-top:14px">${icon("bell")}<span>Enable Developer Mode in Discord, right-click the streamer, and choose <strong>Copy User ID</strong>. Then put <code>{discord_mention}</code> in the message content of any destination template.</span></div>
          <p class="field-hint" style="margin:12px 0 0">Only this user’s mention is allowed. <code>@everyone</code> and role mentions remain blocked.</p>
        </div>
        <div class="modal-foot">
          <button type="button" class="btn btn-ghost" data-action="universal-close">Cancel</button>
          <button type="submit" class="btn btn-primary">${icon("check")} Save ping</button>
        </div>
      </form>
    `, { narrow: true });
  }

  const RESERVED_TWITCH_PATHS = new Set([
    "directory", "videos", "settings", "subscriptions", "wallet", "friends", "inbox",
    "drops", "store", "jobs", "turbo", "search", "login", "signup", "download",
    "prime", "creatorcamp", "legal", "privacy", "terms", "following", "browse",
    "esports", "teams", "about", "careers", "blog", "partners", "affiliates",
    "developers", "help", "support", "security", "p", "embed", "oauth2", "api",
    "collections", "dashboard", "inventory", "messages", "whispers", "channel",
    "video", "clip", "clips", "schedule", "faq", "gift", "redeem", "activate",
    "popout", "moderator", "user", "account", "payments", "purchases", "privacy-center",
  ]);

  function normalizeImportLogin(value) {
    let login = String(value || "").trim().toLowerCase();
    login = login.replace(/^@/, "");
    login = login.replace(/^https?:\/\/(?:www\.)?twitch\.tv\//, "");
    login = login.replace(/^twitch\.tv\//, "");
    login = login.replace(/[/?#].*$/, "");
    login = login.replace(/^#/, "");
    if (!/^[a-z0-9_]{1,25}$/.test(login)) {
      return "";
    }
    return login;
  }

  function collectImportLogins(value, add) {
    if (value == null) {
      return;
    }
    if (typeof value === "string") {
      add(value);
      return;
    }
    if (Array.isArray(value)) {
      value.forEach((item) => collectImportLogins(item, add));
      return;
    }
    if (typeof value === "object") {
      let walkedIdentity = false;
      for (const key of ["login", "channel_login", "username", "user"]) {
        if (typeof value[key] === "string") {
          add(value[key]);
          walkedIdentity = true;
        }
      }
      for (const key of ["logins", "follows", "following", "followed_channels", "channels", "streamers", "data", "items"]) {
        if (value[key]) {
          collectImportLogins(value[key], add);
        }
      }
      if (!walkedIdentity) {
        for (const key of ["name", "display_name", "channel"]) {
          if (typeof value[key] === "string") {
            add(value[key]);
            walkedIdentity = true;
          }
        }
      }
    }
  }

  function parseImportLogins(rawText) {
    const text = String(rawText || "");
    const found = new Set();
    const add = (value) => {
      const login = normalizeImportLogin(value);
      if (login && !RESERVED_TWITCH_PATHS.has(login)) {
        found.add(login);
      }
    };
    const trimmed = text.trim();
    let parsedJSON = false;
    if (trimmed.startsWith("[") || trimmed.startsWith("{")) {
      try {
        collectImportLogins(JSON.parse(trimmed), add);
        parsedJSON = true;
      } catch {
        // Fall back to text scanning below.
      }
    }
    if (!parsedJSON) {
      for (const match of text.matchAll(/twitch\.tv\/([A-Za-z0-9_]{1,25})/gi)) {
        add(match[1]);
      }
      for (const match of text.matchAll(/href=["']\/([A-Za-z0-9_]{1,25})["']/gi)) {
        add(match[1]);
      }
      const looksHTML = /<a\b|<html\b|href\s*=/i.test(text);
      if (!looksHTML) {
        for (const token of text.split(/[\s,;|<>"'`]+/)) {
          add(token);
        }
      }
    }
    return Array.from(found);
  }

  function updateImportParsedCount() {
    const text = $("#import-logins")?.value || "";
    const count = parseImportLogins(text).length;
    const label = $("#import-parsed-count");
    if (label) {
      label.textContent = `${formatCount(count)} username${count === 1 ? "" : "s"}`;
    }
    const submit = $("#import-submit");
    if (submit) {
      submit.disabled = count === 0;
    }
    return count;
  }

  function openImportModal(tab = "helper", prefill = "") {
    importModalTab = tab === "paste" || tab === "signin" ? tab : "helper";
    importFileSource = false;
    const webhooks = state?.config?.webhooks || [];
    const signInEnabled = Boolean(state?.twitch_connect_enabled);
    const autoAssign = webhooks.length === 1;
    openModal(`
      <div class="modal-head">
        <div><h2>Import follows</h2><p>Bring your Twitch follows in without sharing a password or Twitch token.</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="import-form">
        <div class="modal-body">
          <div class="segmented" role="tablist" aria-label="Import method">
            <button type="button" class="segment" role="tab" id="import-tab-helper" aria-controls="import-panel-helper" data-action="import-tab" data-tab="helper">${icon("shield")} No sign-in</button>
            <button type="button" class="segment" role="tab" id="import-tab-paste" aria-controls="import-panel-paste" data-action="import-tab" data-tab="paste">${icon("upload")} Paste a list</button>
            <button type="button" class="segment" role="tab" id="import-tab-signin" aria-controls="import-panel-signin" data-action="import-tab" data-tab="signin">${icon("download")} Sign in to Twitch</button>
          </div>

          <div class="import-panel is-hidden" id="import-panel-helper" role="tabpanel" aria-labelledby="import-tab-helper" data-import-panel="helper">
            <div class="import-note">${icon("shield")}<div><strong>No token ever reaches this app.</strong><span>The helper reads your Twitch following page inside your existing browser session. You copy the list and paste it here.</span></div></div>
            <ol class="helper-steps">
              <li><span>Open the follows helper in a new tab.</span><a class="btn btn-small btn-ghost" href="/follows-helper" target="_blank" rel="noopener">${icon("external")} Open helper</a></li>
              <li><span>Drag the bookmarklet below to your bookmarks bar. You can also copy the code and create a bookmark manually.</span></li>
              <li><span>Open <a href="https://www.twitch.tv/directory/following/channels" target="_blank" rel="noopener">your following page</a>, run the bookmarklet, copy its list, then choose <strong>Paste a list</strong>.</span></li>
            </ol>
            <div class="bookmarklet-box">
              <div class="bookmarklet-row">
                <a id="bookmarklet-link" class="btn btn-primary bookmarklet-link" role="button" aria-disabled="true">${icon("link")} Drag to bookmarks</a>
                <button type="button" class="btn btn-ghost" data-action="copy-bookmarklet">${icon("copy")} Copy code</button>
                <button type="button" class="btn btn-ghost" data-action="import-tab" data-tab="paste">${icon("upload")} Paste instead</button>
              </div>
              <span class="field-hint" id="bookmarklet-status">Loading the bookmarklet…</span>
              <textarea id="bookmarklet-code" class="textarea input-mono import-code" rows="3" readonly placeholder="Bookmarklet code appears here" aria-label="Bookmarklet code"></textarea>
            </div>
            <p class="field-hint">The helper page is served by this app locally. It never sends your follows or a Twitch token to this app — the paste step is the only transfer.</p>
          </div>

          <div class="import-panel is-hidden" id="import-panel-paste" role="tabpanel" aria-labelledby="import-tab-paste" data-import-panel="paste">
            <label class="field">
              <span class="field-label">Paste your follows <span class="counter" id="import-parsed-count">0 usernames</span></span>
              <textarea id="import-logins" class="textarea import-textarea" rows="9" placeholder="streamer_one&#10;streamer_two&#10;https://www.twitch.tv/streamer_three" autocomplete="off" spellcheck="false">${escapeHTML(prefill)}</textarea>
              <span class="field-hint">Plain text, comma-separated lists, Twitch URLs, or saved HTML. Up to 2,000 unique channels. Each channel is validated with Twitch before it is added.</span>
            </label>
            <div class="import-paste-actions">
              <button type="button" class="btn btn-ghost" data-action="pick-import-file">${icon("file")} Choose a file</button>
              <input id="import-file" type="file" accept=".txt,.csv,.json,.html" class="is-hidden">
              <span class="field-hint" id="import-file-name">.txt, .csv, .json, or saved .html</span>
            </div>
            <label class="field">
              <span class="field-label">Assign destinations (optional)</span>
              <div class="destination-checks">
                ${webhooks.length
                  ? webhooks.map((webhook) => `<label class="destination-check"><input type="checkbox" name="import-webhook" value="${escapeHTML(webhook.id)}" ${autoAssign ? "checked" : ""}><span>${escapeHTML(webhook.label)}</span></label>`).join("")
                  : `<span class="field-hint">No destinations yet. Add one after importing, then route the streamers.</span>`}
              </div>
              <span class="field-hint">New streamers post to the selected destinations. Leave all unchecked to auto-assign when there is exactly one destination.</span>
            </label>
            <div id="import-results"></div>
          </div>

          <div class="import-panel is-hidden" id="import-panel-signin" role="tabpanel" aria-labelledby="import-tab-signin" data-import-panel="signin">
            <div class="import-note">${icon("download")}<div><strong>Optional fallback</strong><span>Sign in with a Twitch device code to read the list of channels you follow directly. Tokens stay in memory and are discarded after the import.</span></div></div>
            ${signInEnabled
              ? `<div class="settings-actions"><button type="button" class="btn btn-primary" data-action="start-twitch-import">${icon("download")} Connect Twitch</button></div>`
              : `<div class="warning-box">${icon("bell")}<span>Twitch sign-in is not configured in this build. Use the helper or paste a list instead.</span></div>`}
            <p class="field-hint">The no-sign-in path never asks for Twitch credentials. Use this only if the bookmarklet or paste flow is inconvenient.</p>
          </div>
        </div>
        <div class="modal-foot">
          <span class="field-hint" id="import-foot-hint">Nothing is sent until you paste the list.</span>
          <div class="modal-foot-right">
            <button type="button" class="btn btn-ghost" data-action="close-modal">Cancel</button>
            <button type="submit" class="btn btn-primary is-hidden" id="import-submit">${icon("download")} Import follows</button>
          </div>
        </div>
      </form>
    `, { narrow: true });

    const field = $("#import-logins");
    if (field) {
      field.addEventListener("input", () => {
        importFileSource = false;
        updateImportParsedCount();
      });
    }
    const fileInput = $("#import-file");
    if (fileInput) {
      fileInput.addEventListener("change", async () => {
        const file = fileInput.files?.[0];
        fileInput.value = "";
        if (file) {
          await loadImportFile(file);
        }
      });
    }
    const bookmarkletLink = $("#bookmarklet-link");
    if (bookmarkletLink) {
      bookmarkletLink.addEventListener("click", (event) => {
        event.preventDefault();
        toast("Drag this button to your bookmarks bar, then click it on Twitch.", "info");
      });
    }
    applyImportTab(importModalTab);
  }

  function applyImportTab(tab) {
    importModalTab = tab;
    $$("[data-import-panel]").forEach((panel) => {
      panel.classList.toggle("is-hidden", panel.dataset.importPanel !== tab);
    });
    $$('[data-action="import-tab"]').forEach((button) => {
      const active = button.dataset.tab === tab;
      button.classList.toggle("is-active", active);
      button.setAttribute("aria-selected", active ? "true" : "false");
    });
    const hints = {
      helper: "Nothing is sent until you paste the list.",
      paste: "Already-added streamers are skipped; unresolved entries stay for a retry.",
      signin: "Twitch tokens are discarded after the import.",
    };
    const hint = $("#import-foot-hint");
    if (hint) {
      hint.textContent = hints[tab] || "";
    }
    const submit = $("#import-submit");
    if (submit) {
      submit.classList.toggle("is-hidden", tab !== "paste");
    }
    if (tab === "helper") {
      loadBookmarklet();
    } else if (tab === "paste") {
      updateImportParsedCount();
      setTimeout(() => $("#import-logins")?.focus(), 40);
    }
  }

  function compactBookmarkletSource(source) {
    return String(source || "")
      .replace(/\r/g, "")
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line !== "")
      .join(" ");
  }

  async function loadBookmarklet() {
    const link = $("#bookmarklet-link");
    if (!link) {
      return;
    }
    if (!importBookmarklet && !importBookmarkletPromise) {
      const headers = {};
      const token = storedDashboardToken();
      if (token) {
        headers["X-Dashboard-Token"] = token;
      }
      importBookmarkletPromise = fetch("/static/follows-bookmarklet.js", { cache: "no-store", headers })
        .then((response) => (response.ok ? response.text() : Promise.reject(new Error(`HTTP ${response.status}`))))
        .then((text) => text.trim())
        .catch(() => "");
    }
    if (importBookmarkletPromise) {
      const source = await importBookmarkletPromise;
      if (source) {
        importBookmarklet = `javascript:${compactBookmarkletSource(source.replace(/^javascript:\s*/i, ""))}`;
      } else {
        importBookmarkletPromise = null;
      }
    }
    applyBookmarklet();
  }

  function applyBookmarklet() {
    const link = $("#bookmarklet-link");
    const status = $("#bookmarklet-status");
    const code = $("#bookmarklet-code");
    if (!link) {
      return;
    }
    if (importBookmarklet) {
      link.href = importBookmarklet;
      link.classList.remove("is-disabled");
      link.setAttribute("aria-disabled", "false");
      link.title = "Drag this link to your bookmarks bar, then open it on your Twitch following page.";
      if (status) {
        status.textContent = "Drag this button to your bookmarks bar, then click it on your Twitch following page.";
      }
      if (code) {
        code.value = importBookmarklet;
      }
    } else {
      link.removeAttribute("href");
      link.classList.add("is-disabled");
      link.setAttribute("aria-disabled", "true");
      if (status) {
        status.textContent = "Bookmarklet source is unavailable. Use the helper page or paste a list instead.";
      }
      if (code) {
        code.value = "";
      }
    }
  }

  async function copyBookmarklet() {
    const value = importBookmarklet || $("#bookmarklet-code")?.value || "";
    if (!value) {
      toast("Bookmarklet source is unavailable. Use the helper page or paste a list.", "error");
      return;
    }
    try {
      await navigator.clipboard.writeText(value);
      toast("Bookmarklet code copied.", "success");
    } catch {
      const code = $("#bookmarklet-code");
      if (code) {
        code.focus();
        code.select();
      }
      toast("Could not copy automatically. The code is selected — copy it manually.", "info");
    }
  }

  async function loadImportFile(file) {
    if (!file) {
      return;
    }
    if (file.size > 2 * 1024 * 1024) {
      toast("That file is larger than 2 MB. Paste a smaller list or trim the file.", "error");
      return;
    }
    try {
      const text = await file.text();
      const field = $("#import-logins");
      if (field) {
        field.value = field.value.trim() ? `${field.value.trim()}\n${text}` : text;
      }
      importFileSource = true;
      const name = $("#import-file-name");
      if (name) {
        name.textContent = `${file.name} loaded`;
      }
      updateImportParsedCount();
      toast("File loaded into the paste box.", "success");
    } catch {
      toast("Could not read that file.", "error");
    }
  }

  function renderImportResults(result) {
    const results = $("#import-results");
    if (!results) {
      return;
    }
    const added = Number(result?.added) || 0;
    const assigned = Number(result?.assigned) || 0;
    const skipped = (result?.skipped || []).length;
    const failed = (result?.failed || []).length;
    const unresolved = bulkUnresolved(result);
    results.innerHTML = `
      <div class="import-results">
        <div class="import-summary">
          <span class="badge badge-enabled">${formatCount(added)} added</span>
          <span class="badge">${formatCount(skipped)} skipped</span>
          <span class="badge ${failed ? "badge-invalid" : ""}">${formatCount(failed)} failed</span>
          ${assigned ? `<span class="meta">${formatCount(assigned)} destination assignment${assigned === 1 ? "" : "s"}</span>` : ""}
        </div>
        ${unresolved.length
          ? `<div class="warning-box">${icon("bell")}<span>${formatCount(unresolved.length)} entr${unresolved.length === 1 ? "y needs" : "ies need"} attention. They are back in the paste box so you can retry just those.</span></div>
             <div class="import-issues">${unresolved.slice(0, 12).map((item) => `<div><strong>${escapeHTML(item.login || "unknown")}</strong><span>${escapeHTML(item.reason || "Could not be added")}</span></div>`).join("")}${unresolved.length > 12 ? `<div class="meta">+${unresolved.length - 12} more</div>` : ""}</div>`
          : `<div class="meta">Everything in this batch was handled.</div>`}
      </div>`;
  }

  async function submitImport() {
    const textarea = $("#import-logins");
    const text = textarea?.value.trim() || "";
    if (!text) {
      textarea?.focus();
      toast("Paste at least one Twitch username or channel URL.", "error");
      return;
    }
    if (text.length > 2 * 1024 * 1024) {
      toast("That list is larger than 2 MB. Split it into smaller imports.", "error");
      return;
    }
    const webhookIDs = $$('input[name="import-webhook"]:checked').map((input) => input.value);
    const source = importFileSource ? "file" : "paste";
    const submit = $("#import-submit");
    const results = $("#import-results");
    if (submit) {
      submit.disabled = true;
      submit.innerHTML = '<span class="spinner"></span> Importing…';
    }
    if (results) {
      results.innerHTML = `<div class="import-progress"><span class="spinner dark"></span><span>Validating channels with Twitch…</span></div>`;
    }
    try {
      let response;
      try {
        response = await api("/api/channels/import", {
          method: "POST",
          body: { text, webhook_ids: webhookIDs, source },
        });
      } catch (error) {
        if (error.status === 404 || error.status === 405) {
          // Older builds only expose the bulk endpoint. Keep them working.
          response = await api("/api/channels/bulk", { method: "POST", body: { text } });
        } else {
          throw error;
        }
      }
      if (response.state) {
        state = response.state;
      }
      const result = response.result || {};
      const unresolved = bulkUnresolved(result);
      if (unresolved.length) {
        importFileSource = false;
        if (textarea) {
          textarea.value = unresolved.map((item) => item.login).join("\n");
        }
        renderImportResults(result);
        updateImportParsedCount();
        render();
        toast(bulkResultMessage(result), result.added ? "info" : "error");
      } else {
        closeModal();
        render();
        toast(bulkResultMessage(result), "success");
      }
    } catch (error) {
      if (results) {
        results.innerHTML = `<div class="warning-box">${icon("bell")}<span>${escapeHTML(error.message)}</span></div>`;
      }
      toast(error.message, "error");
    } finally {
      if (submit) {
        submit.innerHTML = `${icon("download")} Import follows`;
        updateImportParsedCount();
      }
    }
  }

  function bulkResultMessage(result, verb = "Added") {
    const added = Number(result?.added) || 0;
    const skipped = (result?.skipped || []).length;
    const failed = (result?.failed || []).length;
    const parts = [`${verb} ${added} streamer${added === 1 ? "" : "s"}.`];
    if (skipped) {
      parts.push(`${skipped} already on your list.`);
    }
    if (failed) {
      parts.push(`${failed} could not be added.`);
    }
    return parts.join(" ");
  }

  function bulkUnresolved(result) {
    const skipped = (result?.skipped || []).filter((item) => item.reason !== "already added");
    return [...skipped, ...(result?.failed || [])];
  }

  async function startTwitchImport() {
    try {
      const view = await api("/api/twitch/imports", { method: "POST", body: {} });
      openModal(renderTwitchImportModal(view), { narrow: true });
      twitchImportState = view;
      bindTwitchImportReadyUI();
      scheduleTwitchImportPoll();
    } catch (error) {
      toast(error.message, "error");
    }
  }

  function renderTwitchImportModal(view) {
    if (view.status === "ready") {
      return renderTwitchImportReady(view);
    }
    const failed = ["expired", "denied", "error"].includes(view.status);
    const minutes = Math.max(1, Math.round((Number(view.expires_in) || 0) / 60));
    return `
      <div class="modal-head">
        <div><h2>${failed ? "Twitch sign-in" : "Connect Twitch"}</h2><p>Import the channels you already follow.</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <div class="modal-body">
        ${failed ? `
          <div class="warning-box">${icon("bell")}<span>${escapeHTML(view.message || "Twitch sign-in did not finish.")}</span></div>
          <p class="field-hint" style="margin-top:12px">Nothing was added to your streamer list.</p>
        ` : `
          <ol class="twitch-steps">
            <li><span>Open the Twitch activation page.</span><a class="btn btn-small btn-primary" href="${escapeHTML(safeURL(view.verification_uri))}" target="_blank" rel="noopener">${icon("external")} Open Twitch</a></li>
            <li><span>Enter this code:</span><span class="device-code">${escapeHTML(view.user_code || "")}<button type="button" class="btn btn-small btn-ghost" data-action="copy-twitch-code" data-code="${escapeHTML(view.user_code || "")}">${icon("copy")} Copy</button></span></li>
            <li><span>Approve access. This window updates by itself.</span></li>
          </ol>
          <div class="twitch-wait"><span class="spinner"></span><span>${escapeHTML(view.message || "Waiting for approval…")}</span></div>
          <p class="field-hint">Only the list of channels you follow is read. Tokens stay in memory and are discarded after the import. The code expires in about ${minutes} minute${minutes === 1 ? "" : "s"}.</p>
        `}
      </div>
      <div class="modal-foot">
        <span class="field-hint">${failed ? "No changes were made." : "You can close this window at any time."}</span>
        <div class="modal-foot-right">
          <button type="button" class="btn btn-ghost" data-action="close-modal">${failed ? "Close" : "Cancel"}</button>
          ${failed ? `<button type="button" class="btn btn-primary" data-action="start-twitch-import">${icon("refresh")} Try again</button>` : ""}
        </div>
      </div>`;
  }

  function renderTwitchImportReady(view) {
    const follows = view.follows || [];
    const addable = follows.filter((follow) => !follow.added);
    const account = view.account || {};
    const rows = follows.map((follow) => {
      const avatar = safeURL(follow.avatar_url || "");
      return `
        <label class="import-row ${follow.added ? "is-added" : ""}" data-login="${escapeHTML(follow.login)}" data-display="${escapeHTML(String(follow.display_name || follow.login).toLowerCase())}">
          <input type="checkbox" name="twitch-follow" value="${escapeHTML(follow.login)}" ${follow.added ? "disabled" : ""}>
          <span class="avatar avatar-sm">${avatar ? `<img src="${escapeHTML(avatar)}" alt="">` : escapeHTML(initials(follow.display_name || follow.login))}</span>
          <span class="import-identity"><strong>${escapeHTML(follow.display_name || follow.login)}</strong><small>${escapeHTML(follow.login)}</small></span>
          ${follow.added ? '<span class="badge badge-paused">Already added</span>' : ""}
        </label>`;
    }).join("");
    return `
      <div class="modal-head">
        <div>
          <h2>Choose streamers</h2>
          <p>${follows.length} followed channel${follows.length === 1 ? "" : "s"} found for ${escapeHTML(account.display_name || account.login || "your account")}. ${addable.length} can be added${addable.length > 1000 ? "; up to 1,000 per import" : ""}.</p>
        </div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="twitch-import-form">
        <div class="modal-body">
          <div class="import-toolbar">
            <input id="twitch-import-search" class="input" placeholder="Filter streamers" aria-label="Filter streamers" autocomplete="off">
            <button type="button" class="btn btn-small btn-ghost" data-action="twitch-select" data-select="pending">Select unadded</button>
            <button type="button" class="btn btn-small btn-ghost" data-action="twitch-select" data-select="none">Clear</button>
          </div>
          <div class="import-list" id="twitch-import-list">${rows}</div>
          ${follows.length ? "" : `<div class="empty-state">${icon("users")}<strong>No follows found</strong><span>This Twitch account does not follow any channels yet.</span></div>`}
        </div>
        <div class="modal-foot">
          <span class="field-hint" id="twitch-import-count">0 selected</span>
          <div class="modal-foot-right">
            <button type="button" class="btn btn-ghost" data-action="close-modal">Cancel</button>
            <button type="submit" class="btn btn-primary" id="twitch-import-submit" disabled>${icon("download")} Import selected</button>
          </div>
        </div>
      </form>`;
  }

  function bindTwitchImportReadyUI() {
    if (twitchImportState?.status !== "ready") {
      return;
    }
    const search = $("#twitch-import-search");
    if (search) {
      search.addEventListener("input", () => {
        const term = search.value.trim().toLowerCase();
        $$(".import-row").forEach((row) => {
          const haystack = `${row.dataset.login || ""} ${row.dataset.display || ""}`;
          row.classList.toggle("is-hidden", term !== "" && !haystack.includes(term));
        });
      });
    }
    const list = $("#twitch-import-list");
    if (list) {
      list.addEventListener("change", updateTwitchImportSelection);
    }
    updateTwitchImportSelection();
  }

  function updateTwitchImportSelection() {
    const checked = $$('input[name="twitch-follow"]:checked').length;
    const count = $("#twitch-import-count");
    if (count) {
      count.textContent = `${checked} selected`;
    }
    const submit = $("#twitch-import-submit");
    if (submit) {
      submit.disabled = checked === 0;
    }
  }

  function setTwitchImportSelection(mode) {
    const inputs = $$('input[name="twitch-follow"]').filter((input) => !input.disabled);
    const limit = 1000;
    inputs.forEach((input, index) => {
      input.checked = mode === "pending" && index < limit;
    });
    if (mode === "pending" && inputs.length > limit) {
      toast(`Selected the first ${limit}. Import them, then connect Twitch again for the rest.`, "info");
    }
    updateTwitchImportSelection();
  }

  function scheduleTwitchImportPoll(delay) {
    clearTimeout(twitchImportTimer);
    if (!twitchImportState?.id) {
      return;
    }
    const seconds = Math.max(3, Number(twitchImportState.interval) || 5);
    twitchImportTimer = setTimeout(pollTwitchImport, delay ?? seconds * 1000);
  }

  async function pollTwitchImport() {
    if (!twitchImportState?.id) {
      return;
    }
    const root = $("#modal-root");
    if (!root || !root.innerHTML) {
      return;
    }
    const id = twitchImportState.id;
    try {
      const view = await api(`/api/twitch/imports/${id}`);
      if (!twitchImportState || twitchImportState.id !== id) {
        return;
      }
      twitchImportState = view;
      const modal = $(".modal", root);
      if (modal) {
        modal.innerHTML = renderTwitchImportModal(view);
      }
      bindTwitchImportReadyUI();
      if (view.status === "pending") {
        scheduleTwitchImportPoll();
      }
    } catch (error) {
      toast(error.message, "error");
    }
  }

function openSampleModal(channelId) {
    const channel = channelById(channelId);
    if (!channel) {
      return;
    }
    const assigned = new Set(assignedWebhookIDs(channel));
    const webhooks = state.config.webhooks || [];
    if (!webhooks.length) {
      toast("Add a Discord destination first.", "error");
      return;
    }
    openModal(`
      <div class="modal-head">
        <div><h2>Send a sample</h2><p>This posts a clearly marked Example using ${escapeHTML(channel.display_name)}'s most recent stream.</p></div>
        <button class="modal-close" data-action="close-modal" aria-label="Close">${icon("x")}</button>
      </div>
      <form id="sample-form" data-channel-id="${channel.id}">
        <div class="modal-body">
          <div class="destination-checks">
            ${webhooks.map((webhook) => `<label class="destination-check"><input type="checkbox" name="sample-webhook" value="${webhook.id}" ${assigned.has(webhook.id) ? "checked" : ""}><span>${escapeHTML(webhook.label)}</span></label>`).join("")}
          </div>
          <p class="field-hint" style="margin:12px 0 0">The message uses each destination's own template and adds an Example marker so your community knows it is a test.</p>
        </div>
        <div class="modal-foot">
          <button type="button" class="btn btn-ghost" data-action="close-modal">Cancel</button>
          <button type="submit" class="btn btn-primary">${icon("send")} Send sample</button>
        </div>
      </form>
    `, { narrow: true });
  }

  function insertPlaceholder(token, field) {
    const target = field || lastFocusedField || $("#tpl-content");
    if (!target) {
      return;
    }
    const start = target.selectionStart ?? target.value.length;
    const end = target.selectionEnd ?? start;
    target.value = `${target.value.slice(0, start)}${token}${target.value.slice(end)}`;
    target.focus();
    target.selectionStart = target.selectionEnd = start + token.length;
    target.dispatchEvent(new Event("input", { bubbles: true }));
  }

  function insertRoutePlaceholder(token) {
    insertPlaceholder(token, routeLastFocusedField || $("#route-content"));
  }

  function exportTemplate() {
    const template = readTemplate();
    const blob = new Blob([JSON.stringify(template, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "twitch-live-template.json";
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
    toast("Template exported.", "success");
  }

  async function importConfig(file) {
    if (!file) {
      return;
    }
    if (!window.confirm("Restore this backup? It replaces your current streamers, destinations, routing, templates, and filters.")) {
      return;
    }
    try {
      const text = await file.text();
      const headers = { "Content-Type": "application/json" };
      const token = storedDashboardToken();
      if (token) {
        headers["X-Dashboard-Token"] = token;
      }
      const response = await fetch("/api/config/import", {
        method: "POST",
        headers,
        body: text,
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(data.error || `Restore failed (${response.status})`);
      }
      state = data;
      applyTheme(state.config.settings.theme);
      render();
      toast("Backup restored. Re-enter any redacted webhook URLs.", "success");
    } catch (error) {
      toast(error.message, "error");
    }
  }

  async function importTemplate(file) {
    if (!file) {
      return;
    }
    try {
      const data = JSON.parse(await file.text());
      fillTemplate({ ...(state.default_template || {}), ...data });
      toast("Template imported into the editor.", "success");
    } catch {
      toast("That file is not a valid template JSON export.", "error");
    }
  }

  async function toggleTheme() {
    if (!state) {
      return;
    }
    const theme = state.config.settings.theme === "dark" ? "light" : "dark";
    state.config.settings.theme = theme;
    applyTheme(theme);
    try {
      state = await api("/api/settings", {
        method: "PUT",
        body: {
          poll_seconds: state.config.settings.poll_seconds,
          start_at_login: state.config.settings.start_at_login,
          paused: state.config.settings.paused,
          disable_update_check: state.config.settings.disable_update_check,
          theme,
        },
      });
      applyTheme(state.config.settings.theme);
      render();
    } catch (error) {
      toast(error.message, "error");
      await refresh(true);
    }
  }

  function toast(message, type = "info") {
    const root = $("#toast-root");
    if (!root) {
      return;
    }
    const item = document.createElement("div");
    item.className = `toast ${type}`;
    item.innerHTML = `<span class="toast-dot"></span><span>${escapeHTML(message)}</span>`;
    root.appendChild(item);
    setTimeout(() => {
      item.remove();
    }, 5200);
  }

  async function triggerCheck() {
    try {
      const result = await api("/api/check", { method: "POST", body: {} });
      if (state) {
        state.checking = Boolean(result.started);
        render();
      }
      toast(result.started ? "Checking every streamer…" : "A check is already running.", "info");
      setTimeout(() => refresh(true), 1200);
    } catch (error) {
      toast(error.message, "error");
    }
  }

  async function handleGlobalClick(event) {
    const actionElement = event.target.closest("[data-action]");
    if (!actionElement) {
      return;
    }
    const action = actionElement.dataset.action;
    switch (action) {
      case "close-modal":
      case "universal-close":
        closeModal();
        break;
      case "go-add-streamer":
        switchView("streamers", true);
        break;
      case "go-history":
        switchView("history");
        break;
      case "go-routing": {
        const routedChannel = (state?.config?.channels || [])[0];
        if (!routedChannel) {
          toast("Add a streamer first, then choose a destination.", "info");
          switchView("streamers", true);
          break;
        }
        streamerView = "matrix";
        switchView("streamers");
        setTimeout(() => {
          $("#streamer-list-section")?.scrollIntoView({ behavior: "smooth", block: "start" });
        }, 80);
        break;
      }
      case "toggle-select-mode":
        streamerSelectMode = !streamerSelectMode;
        if (!streamerSelectMode) {
          selectedStreamerIds.clear();
        }
        syncStreamerToolbarState();
        renderStreamerResults();
        break;
      case "toggle-streamer-view":
        streamerView = streamerView === "matrix" ? "cards" : "matrix";
        syncStreamerToolbarState();
        renderStreamerResults();
        break;
      case "select-visible-streamers":
        selectVisibleStreamers();
        break;
      case "clear-streamer-selection":
        clearStreamerSelection();
        break;
      case "reset-streamer-filters":
        resetStreamerFilters();
        break;
      case "bulk-enable":
        await runBulkAction("enable");
        break;
      case "bulk-disable":
        await runBulkAction("disable");
        break;
      case "bulk-delete":
        await runBulkAction("delete");
        break;
      case "bulk-assign":
        openBulkDestinationModal("assign");
        break;
      case "bulk-unassign":
        openBulkDestinationModal("unassign");
        break;
      case "go-test-webhook": {
        const testDestinations = state?.config?.webhooks || [];
        if (!testDestinations.length) {
          openWebhookModal();
          break;
        }
        switchView("destinations");
        setTimeout(() => {
          $(".destination-card")?.scrollIntoView({ behavior: "smooth", block: "center" });
        }, 80);
        break;
      }
      case "dismiss-onboarding":
        dismissOnboarding();
        break;
      case "open-import":
        openImportModal("helper");
        break;
      case "open-bulk-add":
        openImportModal("paste");
        break;
      case "import-tab":
        applyImportTab(actionElement.dataset.tab || "helper");
        break;
      case "copy-bookmarklet":
        await copyBookmarklet();
        break;
      case "pick-import-file":
        $("#import-file")?.click();
        break;
      case "start-twitch-import":
        await startTwitchImport();
        break;
      case "copy-twitch-code":
        try {
          await navigator.clipboard.writeText(actionElement.dataset.code || "");
          toast("Code copied.", "success");
        } catch {
          toast("Could not copy the code. Select it instead.", "error");
        }
        break;
      case "twitch-select":
        setTwitchImportSelection(actionElement.dataset.select);
        break;
      case "add-destination":
        openWebhookModal();
        break;
      case "edit-webhook":
        openWebhookModal(actionElement.dataset.webhook);
        break;
      case "check-all":
        await triggerCheck();
        break;
      case "check-channel": {
        const channelId = actionElement.dataset.channel;
        try {
          const result = await api(`/api/channels/${channelId}/check`, { method: "POST", body: {} });
          if (state) {
            state.checking = Boolean(result.started);
            render();
          }
          toast(result.started ? "Checking this streamer…" : "A check is already running.", "info");
          setTimeout(() => refresh(true), 1200);
        } catch (error) {
          toast(error.message, "error");
        }
        break;
      }
      case "toggle-theme":
        await toggleTheme();
        break;
      case "toggle-url": {
        const input = $("#wh-url");
        if (input) {
          const showing = input.type === "text";
          input.type = showing ? "password" : "text";
          actionElement.textContent = showing ? "Show" : "Hide";
        }
        break;
      }
      case "reset-template":
        fillTemplate(state.default_template);
        toast("Style reset to the default template.", "info");
        break;
      case "export-template":
        exportTemplate();
        break;
      case "import-template":
        $("#template-import")?.click();
        break;
      case "duplicate-webhook":
        openWebhookModal(actionElement.dataset.webhook, { duplicate: true });
        break;
      case "insert-placeholder":
        insertPlaceholder(actionElement.dataset.token);
        break;
      case "test-webhook": {
        const webhookId = actionElement.dataset.webhook;
        if (!webhookId) {
          break;
        }
        actionElement.disabled = true;
        try {
          await api(`/api/webhooks/${webhookId}/test`, { method: "POST", body: {} });
          toast("Test message posted.", "success");
        } catch (error) {
          toast(error.message, "error");
        } finally {
          actionElement.disabled = false;
        }
        break;
      }
      case "remove-webhook": {
        const webhook = webhookById(actionElement.dataset.webhook);
        if (webhook && window.confirm(`Remove "${webhook.label}"? Streamers connected to it will stop posting there.`)) {
          try {
            state = await api(`/api/webhooks/${webhook.id}`, { method: "DELETE", body: {} });
            render();
            toast("Destination removed.", "info");
          } catch (error) {
            toast(error.message, "error");
          }
        }
        break;
      }
      case "remove-channel": {
        const channel = channelById(actionElement.dataset.channel);
        if (channel && window.confirm(`Remove ${channel.display_name}?`)) {
          try {
            state = await api(`/api/channels/${channel.id}`, { method: "DELETE", body: {} });
            render();
            toast("Streamer removed.", "info");
          } catch (error) {
            toast(error.message, "error");
          }
        }
        break;
      }
      case "toggle-channel": {
        const channel = channelById(actionElement.dataset.channel);
        if (!channel) {
          break;
        }
        actionElement.disabled = true;
        try {
          state = await api(`/api/channels/${channel.id}`, {
            method: "PUT",
            body: { enabled: !channel.enabled },
          });
          render();
        } catch (error) {
          toast(error.message, "error");
          actionElement.disabled = false;
        }
        break;
      }
      case "edit-discord-id":
        openDiscordIDModal(actionElement.dataset.channel);
        break;
      case "edit-filters":
        openFiltersModal(actionElement.dataset.channel);
        break;
      case "edit-route-template":
        event.preventDefault();
        openRouteTemplateModal(actionElement.dataset.channel, actionElement.dataset.webhook);
        break;
      case "clear-route-template":
        await clearRouteTemplate(actionElement.dataset.channel, actionElement.dataset.webhook);
        break;
      case "insert-route-placeholder":
        insertRoutePlaceholder(actionElement.dataset.token);
        break;
      case "sample-channel":
        openSampleModal(actionElement.dataset.channel);
        break;
      case "check-updates":
        actionElement.disabled = true;
        actionElement.innerHTML = '<span class="spinner dark"></span> Checking…';
        try {
          const result = await api("/api/updates/check", { method: "POST", body: {} });
          if (state) {
            state.update = result;
            render();
          }
          if (installJobActive(result.install_job)) {
            updateInstallStartedAt = Date.now();
            scheduleUpdateInstallPoll(700);
          }
          if (result.error) {
            toast(result.error, "error");
          } else if (result.update_available) {
            toast(`Version ${result.latest_version} is available.`, "success");
          } else {
            toast("You are on the latest version.", "success");
          }
        } catch (error) {
          toast(error.message, "error");
          actionElement.disabled = false;
          actionElement.innerHTML = `${icon("refresh")} Check now`;
        }
        break;
      case "install-update":
        await installUpdate();
        break;
      case "copy-diagnostics":
        await copyDiagnostics();
        break;
      case "download-diagnostics":
        await downloadDiagnostics();
        break;
      case "refresh-history":
        refreshHistory();
        break;
      case "clear-history":
        await clearHistory();
        break;
      case "reset-history-filters":
        historyFilters = { kind: "", level: "", channel_id: "", search: "" };
        refreshHistory();
        break;
      case "open-data-folder":
        try {
          await api("/api/open-data-folder", { method: "POST", body: {} });
        } catch (error) {
          toast(error.message, "error");
        }
        break;
      case "export-secrets":
        if (window.confirm("Download a backup with live webhook URLs? Anyone with the file can post to those channels. Keep it private.")) {
          const token = storedDashboardToken();
          const suffix = token ? `&token=${encodeURIComponent(token)}` : "";
          window.location.href = `/api/config/export?secrets=1${suffix}`;
        }
        break;
      case "import-config":
        $("#config-import")?.click();
        break;
      case "clear-dashboard-token":
        if (window.confirm("Remove the dashboard token? Anyone who can reach the dashboard will be able to change settings.")) {
          try {
            state = await api("/api/dashboard-token", { method: "PUT", body: { token: "" } });
            try {
              sessionStorage.removeItem("twitch-live-token");
            } catch {
              // Ignore storage restrictions.
            }
            render();
            toast("Dashboard token removed.", "info");
          } catch (error) {
            toast(error.message, "error");
          }
        }
        break;
      case "quit-app":
        if (!window.confirm("Quit Twitch Live to Discord? Alerts will stop until you start it again.")) {
          break;
        }
        try {
          await api("/api/quit", { method: "POST", body: {} });
          toast("App stopped. You can close this tab.", "success");
          document.body.innerHTML = `<div class="empty-state" style="min-height:100vh;border:0;border-radius:0">${icon("power")}<strong>Twitch Live to Discord has stopped</strong><span>You can close this browser tab. Start the app again whenever you are ready.</span></div>`;
        } catch (error) {
          toast(error.message, "error");
        }
        break;
      default:
        break;
    }
  }

  async function handleChange(event) {
    const target = event.target;
    if (target.id === "streamer-status") {
      streamerStatus = target.value;
      renderStreamerResults();
      return;
    }
    if (target.id === "streamer-sort") {
      streamerSort = target.value;
      renderStreamerResults();
      return;
    }
    if (target.matches('[data-action="toggle-select-channel"]')) {
      toggleStreamerSelection(target);
      return;
    }
    if (target.matches('[data-action="toggle-matrix-cell"]')) {
      await toggleMatrixCell(target);
      return;
    }
    if (target.id === "history-kind" || target.id === "history-level" || target.id === "history-channel") {
      const field = { "history-kind": "kind", "history-level": "level", "history-channel": "channel_id" }[target.id];
      historyFilters[field] = target.value;
      refreshHistory();
      return;
    }
    if (target.id === "template-import") {
      await importTemplate(target.files?.[0]);
      target.value = "";
      return;
    }
    if (target.id === "config-import") {
      await importConfig(target.files?.[0]);
      target.value = "";
      return;
    }
    if (target.id === "tpl-allow-everyone" && target.checked) {
      const confirmed = window.confirm("Allow @everyone / @here for this destination? This can ping the whole server. Continue?");
      if (!confirmed) {
        target.checked = false;
        schedulePreview();
        return;
      }
    }
    if (target.matches('[data-action="toggle-destination"]')) {
      const channel = channelById(target.dataset.channel);
      if (!channel) {
        return;
      }
      const checked = $$(`input[data-action="toggle-destination"][data-channel="${target.dataset.channel}"]`).filter((input) => input.checked).map((input) => input.dataset.webhook);
      try {
        state = await api(`/api/channels/${channel.id}/destinations`, {
          method: "PUT",
          body: { webhook_ids: checked },
        });
        render();
        toast("Destinations updated.", "success");
      } catch (error) {
        toast(error.message, "error");
        await refresh(true);
      }
    }
  }

  function handleInput(event) {
    const target = event.target;
    if (target.id === "streamer-search") {
      streamerSearch = target.value;
      clearTimeout(streamerSearchTimer);
      streamerSearchTimer = setTimeout(() => {
        if (currentView === "streamers") {
          renderStreamerResults();
        }
      }, 140);
      return;
    }
    if (target.id !== "history-search") {
      return;
    }
    const hadFocus = document.activeElement === target;
    const caret = target.selectionStart ?? target.value.length;
    historyFilters.search = target.value;
    clearTimeout(historySearchTimer);
    historySearchTimer = setTimeout(async () => {
      if (currentView !== "history") {
        return;
      }
      historyLoaded = false;
      await loadHistory();
      if (hadFocus) {
        const field = $("#history-search");
        if (field) {
          field.focus();
          const position = Math.min(caret, field.value.length);
          field.setSelectionRange(position, position);
        }
      }
    }, 300);
  }

  async function handleSubmit(event) {
    if (event.target.id === "discord-id-form") {
      event.preventDefault();
      const form = event.target;
      const channelId = form.dataset.channelId;
      const input = $("#discord-user-id");
      const button = $("button[type=submit]", form);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      try {
        state = await api(`/api/channels/${channelId}`, {
          method: "PUT",
          body: { discord_id: input.value.trim() },
        });
        closeModal();
        render();
        toast(input.value.trim() ? "Discord ping saved." : "Discord ping removed.", "success");
      } catch (error) {
        toast(error.message, "error");
        button.disabled = false;
        button.innerHTML = `${icon("check")} Save ping`;
      }
      return;
    }

    if (event.target.id === "import-form") {
      event.preventDefault();
      if (importModalTab !== "paste") {
        return;
      }
      await submitImport();
      return;
    }

    if (event.target.id === "twitch-import-form") {
      event.preventDefault();
      const importID = twitchImportState?.id;
      const logins = $$('input[name="twitch-follow"]:checked').map((input) => input.value);
      if (!importID) {
        toast("This Twitch import expired. Start again.", "error");
        return;
      }
      if (!logins.length) {
        toast("Select at least one streamer.", "error");
        return;
      }
      const button = $("#twitch-import-submit");
      if (button) {
        button.disabled = true;
        button.innerHTML = '<span class="spinner"></span> Importing…';
      }
      try {
        const response = await api(`/api/twitch/imports/${importID}/import`, {
          method: "POST",
          body: { logins },
        });
        twitchImportState = null;
        closeModal();
        state = response.state || state;
        render();
        toast(bulkResultMessage(response.result, "Imported"), "success");
      } catch (error) {
        toast(error.message, "error");
        if (button) {
          button.disabled = false;
          button.innerHTML = `${icon("download")} Import selected`;
        }
      }
      return;
    }

    if (event.target.id === "add-streamer-form") {
      event.preventDefault();
      const form = event.target;
      const input = $("#streamer-login");
      const button = $("button[type=submit]", form);
      const login = input?.value.trim();
      if (!login) {
        input?.focus();
        return;
      }
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Adding…';
      try {
        state = await api("/api/channels", { method: "POST", body: { login } });
        input.value = "";
        currentView = "streamers";
        render();
        toast("Streamer added.", "success");
      } catch (error) {
        toast(error.message, "error");
      } finally {
        button.disabled = false;
        button.innerHTML = `${icon("plus")} Add streamer`;
      }
      return;
    }

    if (event.target.id === "dashboard-token-form") {
      event.preventDefault();
      const button = $("button[type=submit]", event.target);
      const token = $("#dashboard-token")?.value.trim() || "";
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      try {
        state = await api("/api/dashboard-token", { method: "PUT", body: { token } });
        tokenPromptDismissed = false;
        if (token) {
          try {
            sessionStorage.setItem("twitch-live-token", token);
          } catch {
            // Ignore storage restrictions.
          }
        } else {
          try {
            sessionStorage.removeItem("twitch-live-token");
          } catch {
            // Ignore storage restrictions.
          }
        }
        render();
        toast(token ? "Dashboard token saved." : "Dashboard token removed.", "success");
      } catch (error) {
        toast(error.message, "error");
        button.disabled = false;
        button.innerHTML = `${icon("check")} Save token`;
      }
      return;
    }

    if (event.target.id === "settings-form") {
      event.preventDefault();
      const button = $("button[type=submit]", event.target);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      try {
        state = await api("/api/settings", {
          method: "PUT",
          body: {
            poll_seconds: Number($("#settings-interval").value),
            paused: $("#settings-paused").checked,
            start_at_login: $("#settings-startup") ? $("#settings-startup").checked : false,
            disable_update_check: !($("#settings-update-check") ? $("#settings-update-check").checked : true),
            theme: $("#settings-theme").value,
          },
        });
        applyTheme(state.config.settings.theme);
        if (state.update) {
          state.update.disabled = Boolean(state.config.settings.disable_update_check);
        }
        render();
        toast("Settings saved.", "success");
      } catch (error) {
        toast(error.message, "error");
      } finally {
        button.disabled = false;
        button.innerHTML = `${icon("check")} Save settings`;
      }
      return;
    }

    if (event.target.id === "bulk-route-form") {
      event.preventDefault();
      const form = event.target;
      const action = form.dataset.bulkAction === "unassign" ? "unassign" : "assign";
      const ids = $$('input[name="bulk-webhook"]:checked', form).map((input) => input.value);
      if (!ids.length) {
        toast("Choose at least one destination.", "error");
        return;
      }
      const button = $("button[type=submit]", form);
      const label = button ? button.innerHTML : "";
      if (button) {
        button.disabled = true;
        button.innerHTML = '<span class="spinner"></span> Saving…';
      }
      try {
        await runBulkAction(action, ids);
      } finally {
        if (button?.isConnected) {
          button.disabled = false;
          button.innerHTML = label;
        }
      }
      return;
    }

    if (event.target.id === "route-template-form") {
      event.preventDefault();
      const form = event.target;
      const channelId = form.dataset.channelId;
      const webhookId = form.dataset.webhookId;
      const button = $("button[type=submit]", form);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      try {
        state = await api(`/api/channels/${channelId}/destinations/${webhookId}`, {
          method: "PUT",
          body: { template: readRouteTemplate() },
        });
        closeModal();
        render();
        toast("Route template saved.", "success");
      } catch (error) {
        toast(error.message, "error");
        button.disabled = false;
        button.innerHTML = `${icon("check")} Save route template`;
      }
      return;
    }

    if (event.target.id === "filters-form") {
      event.preventDefault();
      const form = event.target;
      const channelId = form.dataset.channelId;
      const button = $("button[type=submit]", form);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      try {
        state = await api(`/api/channels/${channelId}`, {
          method: "PUT",
          body: { filters: readFilters() },
        });
        closeModal();
        render();
        toast("Alert filters saved.", "success");
      } catch (error) {
        toast(error.message, "error");
        button.disabled = false;
        button.innerHTML = `${icon("check")} Save filters`;
      }
      return;
    }

    if (event.target.id === "webhook-form") {
      event.preventDefault();
      const form = event.target;
      const id = form.dataset.webhookId || "";
      const button = $("button[type=submit]", form);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      const template = readTemplate();
      const cleanupMode = $("#wh-cleanup-mode")?.value || "off";
      const cleanup = { mode: cleanupMode };
      if (cleanupMode === "edit" || cleanupMode === "post") {
        cleanup.template = readEndedTemplate();
      }
      const body = {
        label: $("#wh-label").value.trim(),
        url: $("#wh-url").value.trim(),
        role_ids: parseRoleIDs($("#wh-roles")?.value || ""),
        enabled: $("#wh-enabled").checked,
        template,
        username: $("#wh-username")?.value.trim() || "",
        avatar_url: $("#wh-avatar")?.value.trim() || "",
        thread_id: $("#wh-thread")?.value.trim() || "",
        buttons: parseButtons($("#wh-buttons")?.value || ""),
        update_seconds: Number($("#wh-update-seconds")?.value) || 0,
        cleanup,
      };
      try {
        state = await api(id ? `/api/webhooks/${id}` : "/api/webhooks", {
          method: id ? "PUT" : "POST",
          body,
        });
        closeModal();
        render();
        toast(id ? "Destination saved." : "Destination added.", "success");
      } catch (error) {
        toast(error.message, "error");
        button.disabled = false;
        button.innerHTML = `${icon("check")} ${id ? "Save destination" : "Add destination"}`;
      }
      return;
    }

    if (event.target.id === "sample-form") {
      event.preventDefault();
      const form = event.target;
      const channelId = form.dataset.channelId;
      const ids = $$('input[name="sample-webhook"]:checked', form).map((input) => input.value);
      if (!ids.length) {
        toast("Choose at least one destination.", "error");
        return;
      }
      const button = $("button[type=submit]", form);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Sending…';
      try {
        const result = await api(`/api/channels/${channelId}/sample`, {
          method: "POST",
          body: { webhook_ids: ids },
        });
        closeModal();
        await refresh(true);
        toast(result.message || "Sample posted.", "success");
      } catch (error) {
        toast(error.message, "error");
        button.disabled = false;
        button.innerHTML = `${icon("send")} Send sample`;
      }
    }
  }

  async function init() {
    try {
      const savedTheme = localStorage.getItem("twitch-live-theme");
      if (savedTheme) {
        applyTheme(savedTheme);
      }
    } catch {
      // Ignore storage restrictions.
    }
    document.addEventListener("click", handleGlobalClick);
    document.addEventListener("change", handleChange);
    document.addEventListener("input", handleInput);
    document.addEventListener("submit", handleSubmit);
    $("#theme-toggle").addEventListener("click", toggleTheme);
    $$(".nav-item").forEach((item) => {
      item.addEventListener("click", () => switchView(item.dataset.view));
    });
    const knownViews = ["overview", "streamers", "destinations", "history", "settings"];
    const hashView = location.hash.replace("#", "");
    if (knownViews.includes(hashView)) {
      currentView = hashView;
    }
    window.addEventListener("hashchange", () => {
      const view = location.hash.replace("#", "");
      if (knownViews.includes(view)) {
        switchView(view);
      }
    });
    await refresh();
    if (!state) {
      toast("Could not reach the local app. Is it still running?", "error");
    }
    setTimeout(poll, 1200);
  }

  init();
})();
