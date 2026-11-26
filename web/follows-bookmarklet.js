(function () {
  'use strict';

  var RESERVED = {
    about: 1, activate: 1, app: 1, apps: 1, auth: 1, browse: 1, careers: 1,
    channel: 1, channels: 1, collections: 1, communities: 1, contact: 1,
    creatorcamp: 1, dashboard: 1, directory: 1, download: 1, downloads: 1,
    drops: 1, embed: 1, esports: 1, following: 1, friends: 1, gift: 1,
    help: 1, home: 1, index: 1, inventory: 1, jobs: 1, legal: 1, login: 1,
    logout: 1, messages: 1, mobile: 1, moderation: 1, notifications: 1,
    p: 1, partner: 1, player: 1, popout: 1, preferences: 1, prime: 1,
    privacy: 1, search: 1, settings: 1, signup: 1, store: 1, subscriptions: 1,
    support: 1, team: 1, teams: 1, tos: 1, turbo: 1, videos: 1, wallet: 1,
    watch: 1
  };
  var MAX_SCROLLS = 40;
  var SETTLE_ROUNDS = 3;
  var PANEL_ID = 'tln-follows-export';

  function sleep(ms) {
    return new Promise(function (resolve) { setTimeout(resolve, ms); });
  }

  function loginFromHref(href) {
    if (!href) {
      return '';
    }
    var value = String(href).trim();
    if (!value || value.charAt(0) === '#') {
      return '';
    }
    var lower = value.toLowerCase();
    var path = '';
    if (lower.indexOf('http://') === 0 || lower.indexOf('https://') === 0 || value.indexOf('//') === 0) {
      try {
        var url = new URL(value.indexOf('//') === 0 ? 'https:' + value : value, window.location.href);
        var host = url.hostname.toLowerCase();
        if (host !== 'twitch.tv' && host.slice(-10) !== '.twitch.tv') {
          return '';
        }
        path = url.pathname;
      } catch (error) {
        return '';
      }
    } else if (value.charAt(0) === '/') {
      path = value;
    } else {
      return '';
    }

    path = path.replace(/^\/+/, '');
    var cut = path.search(/[?#]/);
    if (cut >= 0) {
      path = path.slice(0, cut);
    }
    path = path.replace(/\/+$/, '');
    if (!path || path.indexOf('/') >= 0) {
      return '';
    }
    try {
      path = decodeURIComponent(path);
    } catch (error) {
      return '';
    }
    var login = path.toLowerCase();
    if (!/^[a-z0-9_]{1,25}$/.test(login) || RESERVED[login]) {
      return '';
    }
    return login;
  }

  function collectLogins() {
    var seen = {};
    var logins = [];
    var anchors = document.querySelectorAll('a[href]');
    for (var index = 0; index < anchors.length; index++) {
      var login = loginFromHref(anchors[index].getAttribute('href'));
      if (login && !seen[login]) {
        seen[login] = true;
        logins.push(login);
      }
    }
    return logins;
  }

  async function loadAll() {
    var scrollRoot = document.scrollingElement || document.documentElement;
    var previous = -1;
    var stable = 0;
    for (var index = 0; index < MAX_SCROLLS; index++) {
      var showMore = document.querySelectorAll('button');
      for (var buttonIndex = 0; buttonIndex < showMore.length; buttonIndex++) {
        var text = (showMore[buttonIndex].textContent || '').trim().toLowerCase();
        if (text === 'show more' || text === 'load more' || text === 'see more') {
          showMore[buttonIndex].click();
        }
      }
      window.scrollTo(0, scrollRoot.scrollHeight);
      await sleep(450);
      var count = collectLogins().length;
      if (count === previous) {
        stable++;
        if (stable >= SETTLE_ROUNDS) {
          break;
        }
      } else {
        stable = 0;
        previous = count;
      }
    }
    window.scrollTo(0, 0);
    return collectLogins();
  }

  function removeOverlay() {
    var existing = document.getElementById(PANEL_ID);
    if (existing && existing.parentNode) {
      existing.parentNode.removeChild(existing);
    }
  }

  function legacyCopy(text) {
    var field = document.createElement('textarea');
    field.value = text;
    field.setAttribute('readonly', '');
    field.style.position = 'fixed';
    field.style.top = '-1000px';
    field.style.left = '-1000px';
    document.body.appendChild(field);
    field.select();
    var copied = false;
    try {
      copied = document.execCommand('copy');
    } catch (error) {
      copied = false;
    }
    document.body.removeChild(field);
    return copied;
  }

  function copyText(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(function () {
        return true;
      }, function () {
        return legacyCopy(text);
      });
    }
    return Promise.resolve(legacyCopy(text));
  }

  function buildOverlay(message) {
    removeOverlay();
    var panel = document.createElement('div');
    panel.id = PANEL_ID;
    panel.style.cssText = 'position:fixed;right:20px;bottom:20px;z-index:2147483647;width:min(430px,calc(100vw - 40px));background:#18181b;color:#efeff1;border:1px solid #3a3a3d;border-radius:12px;box-shadow:0 16px 48px rgba(0,0,0,.6);font:13px/1.45 Inter,system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;overflow:hidden;';

    var header = document.createElement('div');
    header.style.cssText = 'display:flex;align-items:center;justify-content:space-between;gap:12px;padding:12px 14px;background:#202024;border-bottom:1px solid #3a3a3d;';
    var title = document.createElement('strong');
    title.textContent = 'Twitch follows export';
    title.style.cssText = 'font-size:14px;';
    var close = document.createElement('button');
    close.type = 'button';
    close.textContent = 'Close';
    close.style.cssText = 'background:transparent;color:#adadb8;border:0;cursor:pointer;font:inherit;padding:2px 4px;';
    close.onclick = removeOverlay;
    header.appendChild(title);
    header.appendChild(close);
    panel.appendChild(header);

    var body = document.createElement('div');
    body.style.cssText = 'padding:12px 14px 14px;';
    var note = document.createElement('p');
    note.textContent = 'Copy the list, then paste it into the notifier. Nothing is uploaded or sent anywhere from this page.';
    note.style.cssText = 'margin:0 0 10px;color:#adadb8;';
    body.appendChild(note);

    var textarea = document.createElement('textarea');
    textarea.value = message || '';
    textarea.spellcheck = false;
    textarea.setAttribute('aria-label', 'Followed Twitch channel logins');
    textarea.style.cssText = 'box-sizing:border-box;width:100%;height:190px;resize:vertical;background:#0e0e10;color:#efeff1;border:1px solid #3a3a3d;border-radius:8px;padding:9px 10px;font:12px/1.5 ui-monospace,SFMono-Regular,Consolas,monospace;';
    body.appendChild(textarea);

    var footer = document.createElement('div');
    footer.style.cssText = 'display:flex;align-items:center;gap:10px;margin-top:10px;';
    var copy = document.createElement('button');
    copy.type = 'button';
    copy.textContent = 'Copy list';
    copy.style.cssText = 'background:#9147ff;color:#fff;border:0;border-radius:6px;padding:7px 13px;cursor:pointer;font:600 13px/1.2 inherit;';
    var status = document.createElement('span');
    status.style.cssText = 'color:#adadb8;font-size:12px;';
    var count = function () {
      var lines = textarea.value.split(/\r?\n/).filter(function (line) { return line.trim() !== ''; });
      return lines.length;
    };
    status.textContent = count() + ' username' + (count() === 1 ? '' : 's');
    textarea.addEventListener('input', function () {
      status.textContent = count() + ' username' + (count() === 1 ? '' : 's');
    });
    copy.onclick = function () {
      copyText(textarea.value).then(function (ok) {
        status.textContent = ok ? 'Copied ' + count() + ' username' + (count() === 1 ? '' : 's') + '.' : 'Copy failed. Select the text and copy it manually.';
      });
    };
    footer.appendChild(copy);
    footer.appendChild(status);
    body.appendChild(footer);
    panel.appendChild(body);

    document.body.appendChild(panel);
    textarea.focus();
    textarea.select();
    return { textarea: textarea, status: status };
  }

  async function run() {
    var overlay = buildOverlay('Loading your followed channels…');
    overlay.textarea.readOnly = true;
    try {
      var logins = await loadAll();
      var text = logins.join('\n');
      buildOverlay(text);
      if (!logins.length) {
        var panel = document.getElementById(PANEL_ID);
        var note = panel ? panel.querySelector('p') : null;
        if (note) {
          note.textContent = 'No channel links were found. Open your following page, scroll once, and run the bookmarklet again.';
        }
      }
    } catch (error) {
      buildOverlay('');
      var current = document.getElementById(PANEL_ID);
      var message = current ? current.querySelector('p') : null;
      if (message) {
        message.textContent = 'Could not read this page automatically. Select the channel links manually or use the paste import instead.';
      }
    }
  }

  run();
})();
