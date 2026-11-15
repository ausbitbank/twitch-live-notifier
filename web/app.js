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
  let tokenPromptActive = false;
  let tokenPromptDismissed = false;

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
      throw new Error(data.error || `Request failed (${response.status})`);
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
      applyTheme(state.config.settings.theme);
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
    }
    $("#top-actions").innerHTML = actions.join("");
  }

  function renderOverview() {
    const channels = state.config.channels || [];
    const webhooks = state.config.webhooks || [];
    const live = liveChannels();
    const connectionCount = channels.reduce((total, channel) => total + assignedWebhookIDs(channel).length, 0);
    const allAssigned = channels.length > 0 && channels.every((channel) => assignedWebhookIDs(channel).length > 0);

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

    const onboarding = renderOnboarding(channels, webhooks, allAssigned);
    const liveSection = live.length
      ? `<div class="live-list">${live.map(renderLiveCard).join("")}</div>`
      : `<div class="empty-state">${icon("play")}<strong>No one is live right now</strong><span>The dashboard refreshes automatically.</span></div>`;
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

  function renderOnboarding(channels, webhooks, allAssigned) {
    const steps = [
      [webhooks.length > 0, "Add a Discord destination"],
      [channels.length > 0, "Add a Twitch streamer"],
      [allAssigned, "Choose where alerts go"],
      [!state.config.settings.paused && channels.length > 0, "Leave monitoring running"],
    ];
    if (steps.every(([done]) => done)) {
      return "";
    }
    return `
      <section class="card section">
        <div class="card-head">
          <div><h3>Finish setup</h3><p>Four small steps and your community is covered.</p></div>
          <span class="badge">${steps.filter(([done]) => done).length}/4</span>
        </div>
        <div class="card-body">
          <div class="step-list">
            ${steps.map(([done, label]) => `<div class="step ${done ? "is-done" : ""}"><span class="step-check">${done ? "✓" : ""}</span><span>${escapeHTML(label)}</span></div>`).join("")}
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
      return `<div class="empty-state">${icon("bell")}<strong>Quiet for now</strong><span>Alerts will show up here.</span></div>`;
    }
    return `<div class="activity-list">${activity.map((item) => `
      <div class="activity-item">
        <span class="activity-dot ${escapeHTML(item.level)}"></span>
        <div class="activity-text"><p>${escapeHTML(item.message)}</p><span class="activity-time">${escapeHTML(formatRelative(item.time))}</span></div>
      </div>`).join("")}</div>`;
  }

  function renderStreamers() {
    const channels = state.config.channels || [];
    const webhooks = state.config.webhooks || [];
    const addForm = `
      <section class="card section">
        <div class="card-head">
          <div><h3>Add a streamer</h3><p>Paste a Twitch username or a full channel URL.</p></div>
        </div>
        <div class="card-body">
          <form id="add-streamer-form" class="form-row" autocomplete="off">
            <input id="streamer-login" class="input" name="login" placeholder="twitch.tv/your_channel" aria-label="Twitch username or URL">
            <button class="btn btn-primary" type="submit" style="flex:none">${icon("plus")} Add streamer</button>
          </form>
          <p class="field-hint" style="margin:9px 0 0">No limit. Add every channel your community cares about.</p>
        </div>
      </section>`;

    const list = channels.length
      ? channels.map((channel) => renderChannelCard(channel, webhooks)).join("")
      : `<div class="empty-state section">${icon("users")}<strong>No streamers yet</strong><span>Add a Twitch channel above to get started.</span></div>`;

    $("#view-streamers").innerHTML = `${addForm}<section class="section"><div class="section-head"><div><h2>Your streamers</h2><p>${channels.length} configured</p></div></div>${list}</section>`;
  }

  function renderChannelCard(channel, webhooks) {
    const status = statusFor(channel.id);
    const assigned = new Set(assignedWebhookIDs(channel));
    const badge = !channel.enabled
      ? '<span class="badge badge-paused">Paused</span>'
      : status.live
        ? '<span class="badge badge-live">Live</span>'
        : '<span class="badge badge-offline">Offline</span>';

    const destinationControls = webhooks.length
      ? `<div class="destination-checks">${webhooks.map((webhook) => {
          const destination = (channel.destinations || []).find((item) => item.webhook_id === webhook.id);
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
      <article class="channel-card">
        <div class="channel-head">
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

  function renderDestinations() {
    const webhooks = state.config.webhooks || [];
    const cards = webhooks.length
      ? webhooks.map((webhook) => renderDestinationCard(webhook)).join("")
      : `<div class="empty-state">${icon("send")}<strong>No Discord destinations yet</strong><span>Add a webhook for each Discord channel or server you want to post into.</span></div>`;
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

  function renderUpdateStatus(update) {
    const current = escapeHTML(update.current_version || state.version);
    if (update.disabled) {
      return { html: `<div class="meta">Update checks are off. You are on version <strong>${current}</strong>.</div>` };
    }
    if (update.error) {
      return { html: `<div class="warning-box">${icon("bell")} <span>Could not check for updates: ${escapeHTML(update.error)}</span></div>` };
    }
    if (update.update_available) {
      return { html: `<div class="warning-box">${icon("download")} <span>Version <strong>${escapeHTML(update.latest_version)}</strong> is available. You are on <strong>${current}</strong>. Installed settings are preserved when you upgrade.</span></div>` };
    }
    if (update.latest_version) {
      return { html: `<div class="meta">You are on the latest version, <strong>${current}</strong>${update.checked_at ? ` · checked ${escapeHTML(formatRelative(update.checked_at))}` : ""}.</div>` };
    }
    return { html: `<div class="meta">You are on version <strong>${current}</strong>. No update check has run yet.</div>` };
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
                  <div class="switch-copy"><strong>Check GitHub for updates</strong><span>The app never auto-downloads a release. It only reports the latest version.</span></div>
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
            <div class="card-head"><div><h3>Updates</h3><p>Version status from GitHub releases.</p></div></div>
            <div class="card-body">
              <div class="update-status">${updateStatus.html}</div>
              <div class="settings-actions">
                <button class="btn btn-ghost" data-action="check-updates" ${update.disabled ? "disabled" : ""}>${icon("refresh")} Check now</button>
                ${update.update_available && update.download_url ? `<a class="btn btn-primary" href="${escapeHTML(safeURL(update.download_url))}" target="_blank" rel="noopener">${icon("download")} Download ${escapeHTML(update.latest_version || "update")}</a>` : ""}
                ${update.release_url ? `<a class="btn btn-ghost" href="${escapeHTML(safeURL(update.release_url))}" target="_blank" rel="noopener">${icon("external")} Release notes</a>` : ""}
              </div>
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
    syncEndedFields();
    updateCounters();
    schedulePreview();
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
    document.addEventListener("submit", handleSubmit);
    $("#theme-toggle").addEventListener("click", toggleTheme);
    $$(".nav-item").forEach((item) => {
      item.addEventListener("click", () => switchView(item.dataset.view));
    });
    const hashView = location.hash.replace("#", "");
    if (["overview", "streamers", "destinations", "settings"].includes(hashView)) {
      currentView = hashView;
    }
    window.addEventListener("hashchange", () => {
      const view = location.hash.replace("#", "");
      if (["overview", "streamers", "destinations", "settings"].includes(view)) {
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
