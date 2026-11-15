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
  let escHandler = null;

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

  async function api(path, { method = "GET", body } = {}) {
    const options = { method, headers: {}, cache: "no-store" };
    if (method !== "GET") {
      options.headers["Content-Type"] = "application/json";
      options.body = JSON.stringify(body ?? {});
    }
    const response = await fetch(path, options);
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
      ? `<div class="destination-checks">${webhooks.map((webhook) => `
          <label class="destination-check ${webhook.enabled ? "" : "is-disabled"}">
            <input type="checkbox" data-action="toggle-destination" data-channel="${channel.id}" data-webhook="${webhook.id}" ${assigned.has(webhook.id) ? "checked" : ""}>
            <span>${escapeHTML(webhook.label)}</span>
            ${webhook.enabled ? "" : '<span class="badge badge-disabled">Paused</span>'}
          </label>`).join("")}</div>`
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
    return `
      <article class="destination-card">
        <div class="destination-head">
          <div class="destination-icon">${icon("send")}</div>
          <div class="destination-copy">
            <strong>${escapeHTML(webhook.label)}</strong>
            <code>${escapeHTML(maskWebhook(webhook.url))}</code>
          </div>
          <span class="badge ${webhook.enabled ? "badge-enabled" : "badge-disabled"}">${webhook.enabled ? "Active" : "Paused"}</span>
        </div>
        <div class="template-preview-line"><strong>Message:</strong> ${escapeHTML(summary)}</div>
        <div class="meta" style="margin-top:8px">${connected} streamer${connected === 1 ? "" : "s"} connected</div>
        <div class="destination-actions">
          <button class="btn btn-small btn-ghost" data-action="edit-webhook" data-webhook="${webhook.id}">${icon("edit")} Edit & preview</button>
          <button class="btn btn-small btn-ghost" data-action="test-webhook" data-webhook="${webhook.id}">${icon("send")} Test webhook</button>
          <button class="btn btn-small btn-danger" data-action="remove-webhook" data-webhook="${webhook.id}">${icon("trash")} Remove</button>
        </div>
      </article>`;
  }

  function renderSettings() {
    const settings = state.config.settings;
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
              <p class="field-hint" style="margin:9px 0 0">Webhook URLs are stored in this file. Keep it private.</p>
              <div class="settings-actions">${windowsTools}<button class="btn btn-danger" data-action="quit-app">${icon("power")} Quit app</button></div>
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
    previewSequence++;
    lastFocusedField = null;
  }

  function openWebhookModal(webhookId = "") {
    const existing = webhookId ? webhookById(webhookId) : null;
    const template = { ...(state.default_template || {}), ...(existing?.template || {}) };
    const channels = state.config.channels || [];
    const previewOptions = channels.length
      ? channels.map((channel) => `<option value="${channel.id}">${escapeHTML(channel.display_name || channel.login)}</option>`).join("")
      : '<option value="">Sample stream</option>';
    const modalTitle = existing ? "Edit destination" : "New destination";
    const modalSubtitle = existing
      ? "Change the label, webhook, or message style."
      : "Pick a Discord webhook and make the alert your own.";
    const enabled = existing ? existing.enabled : true;

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
                <input id="wh-label" class="input" value="${escapeHTML(existing?.label || "")}" placeholder="Community #live" autocomplete="off">
                <span class="field-hint">This is what you see when choosing where alerts go.</span>
              </label>
              <label class="field">
                <span class="field-label">Discord webhook URL <button type="button" class="placeholder-chip" data-action="toggle-url">Show</button></span>
                <input id="wh-url" class="input input-mono" type="password" value="${escapeHTML(existing?.url || "")}" placeholder="https://discord.com/api/webhooks/..." autocomplete="off" spellcheck="false">
                <span class="field-hint">Server Settings → Integrations → Webhooks → New Webhook → Copy URL.</span>
              </label>
              <div class="switch-row">
                <div class="switch-copy"><strong>Destination active</strong><span>When paused, alerts are not sent here.</span></div>
                <label class="switch"><input id="wh-enabled" type="checkbox" ${enabled ? "checked" : ""}><span class="switch-track"></span></label>
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
    const colorText = $("#tpl-color-text");
    const colorPicker = $("#tpl-color");
    if (colorText && colorPicker) {
      colorText.addEventListener("input", () => {
        if (/^#[0-9a-fA-F]{6}$/.test(colorText.value)) {
          colorPicker.value = colorText.value;
        }
      });
      colorPicker.addEventListener("input", () => {
        colorText.value = colorPicker.value;
      });
    }
    updateCounters();
    schedulePreview();
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
    };
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
    const toggles = {
      "tpl-author": template.show_author,
      "tpl-image": template.show_image,
      "tpl-game": template.show_game,
      "tpl-viewers": template.show_viewers,
      "tpl-live": template.show_live_since,
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
    const template = readTemplate();
    const channelId = $("#preview-channel")?.value || "";
    const webhookLabel = $("#wh-label")?.value || "Discord";
    try {
      const data = await api("/api/preview", {
        method: "POST",
        body: {
          template,
          channel_id: channelId,
          webhook_label: webhookLabel,
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
      const usesDiscordMention = [template.content, template.embed_title, template.embed_description, template.footer]
        .some((value) => String(value).includes("{discord_mention}"));
      const missingDiscordID = usesDiscordMention && channelId && previewChannel && !previewChannel.discord_id;
      if (unknown.length) {
        warning.innerHTML = `<div class="warning-box" style="margin-top:10px">${icon("bell")} Unknown placeholder${unknown.length === 1 ? "" : "s"}: ${escapeHTML(unknown.join(", "))}</div>`;
      } else if (missingDiscordID) {
        warning.innerHTML = `<div class="warning-box" style="margin-top:10px">${icon("bell")} This streamer has no Discord user ID yet. Add one under Streamers, or {discord_mention} will render as plain @text.</div>`;
      } else if (empty) {
        warning.innerHTML = `<div class="warning-box" style="margin-top:10px">${icon("bell")} Nothing will be posted. Add message content or embed text.</div>`;
      } else {
        warning.innerHTML = "";
      }
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
    return `
      <div class="discord-message">
        <div class="discord-avatar"><img src="/static/icon.png" alt=""></div>
        <div class="discord-message-body">
          <div class="discord-message-head"><span class="discord-author">Twitch Live</span><span class="discord-bot-tag">App</span><span class="discord-timestamp">Today at ${escapeHTML(timestamp)}</span></div>
          ${content ? `<div class="discord-content">${discordMarkdown(content)}</div>` : ""}
          ${embed ? `<div class="discord-embed" style="border-left-color:${color}">
            ${author.name ? `<div class="discord-embed-author">${authorIcon ? `<img src="${escapeHTML(authorIcon)}" alt="">` : ""}${escapeHTML(author.name)}</div>` : ""}
            ${title ? `<div class="discord-embed-title">${title}</div>` : ""}
            ${embed.description ? `<div class="discord-embed-desc">${discordMarkdown(embed.description)}</div>` : ""}
            ${fields ? `<div class="discord-fields">${fields}</div>` : ""}
            ${imageHTML || (showImage ? `<div class="discord-empty-image">Twitch preview appears here</div>` : "")}
            ${footer}
          </div>` : ""}
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
    text = text.replace(/&lt;@!?(\d+)&gt;/g, '<span class="discord-mention">@mention</span>');
    return text.replace(/\n/g, "<br>");
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

  function insertPlaceholder(token) {
    const field = lastFocusedField || $("#tpl-content");
    if (!field) {
      return;
    }
    const start = field.selectionStart ?? field.value.length;
    const end = field.selectionEnd ?? start;
    field.value = `${field.value.slice(0, start)}${token}${field.value.slice(end)}`;
    field.focus();
    field.selectionStart = field.selectionEnd = start + token.length;
    field.dispatchEvent(new Event("input", { bubbles: true }));
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
      case "sample-channel":
        openSampleModal(actionElement.dataset.channel);
        break;
      case "open-data-folder":
        try {
          await api("/api/open-data-folder", { method: "POST", body: {} });
        } catch (error) {
          toast(error.message, "error");
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
            theme: $("#settings-theme").value,
          },
        });
        applyTheme(state.config.settings.theme);
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

    if (event.target.id === "webhook-form") {
      event.preventDefault();
      const form = event.target;
      const id = form.dataset.webhookId || "";
      const button = $("button[type=submit]", form);
      button.disabled = true;
      button.innerHTML = '<span class="spinner"></span> Saving…';
      const template = readTemplate();
      const body = {
        label: $("#wh-label").value.trim(),
        url: $("#wh-url").value.trim(),
        enabled: $("#wh-enabled").checked,
        template,
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
