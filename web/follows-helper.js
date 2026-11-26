(function () {
  'use strict';

  var bookmarklet = document.getElementById('bookmarklet');
  var status = document.getElementById('bookmarklet-status');
  var codeArea = document.getElementById('bookmarklet-code');
  var copyButton = document.getElementById('copy-code');
  var consoleButton = document.getElementById('copy-console');
  var source = '';

  function compact(sourceText) {
    return sourceText
      .replace(/\r/g, '')
      .split('\n')
      .map(function (line) { return line.trim(); })
      .filter(function (line) { return line !== ''; })
      .join(' ');
  }

  function setStatus(message) {
    status.textContent = message;
  }

  function legacyCopy(text) {
    codeArea.focus();
    codeArea.select();
    var copied = false;
    try {
      copied = document.execCommand('copy');
    } catch (error) {
      copied = false;
    }
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

  fetch('/static/follows-bookmarklet.js', { cache: 'no-store' })
    .then(function (response) {
      if (!response.ok) {
        throw new Error('HTTP ' + response.status);
      }
      return response.text();
    })
    .then(function (text) {
      source = 'javascript:' + compact(text);
      bookmarklet.setAttribute('href', source);
      bookmarklet.removeAttribute('aria-disabled');
      codeArea.value = source;
      setStatus('Ready. Drag the button to your bookmarks bar.');
    })
    .catch(function () {
      bookmarklet.setAttribute('href', '#');
      bookmarklet.setAttribute('aria-disabled', 'true');
      setStatus('Could not load the bookmarklet. Use the manual copy below.');
    });

  copyButton.addEventListener('click', function () {
    if (!source) {
      setStatus('The bookmarklet is not loaded yet. Reload this page and try again.');
      return;
    }
    copyText(source).then(function (ok) {
      if (ok) {
        setStatus('Bookmarklet code copied. Paste it as a bookmark URL.');
      } else {
        codeArea.focus();
        codeArea.select();
        setStatus('Copy failed. Select the code and copy it manually.');
      }
    });
  });

  if (consoleButton) {
    consoleButton.addEventListener('click', function () {
      if (!source) {
        setStatus('The script is not loaded yet. Reload this page and try again.');
        return;
      }
      var script = source.replace(/^javascript:\s*/i, '');
      copyText(script).then(function (ok) {
        if (ok) {
          setStatus('Console script copied. Paste it into DevTools → Console on the Twitch following page.');
        } else {
          codeArea.focus();
          codeArea.select();
          setStatus('Copy failed. Select the code, remove the javascript: prefix, and copy it manually.');
        }
      });
    });
  }

  // Clicking a javascript: bookmarklet while this local helper page is open
  // would run it in the helper's context, not on Twitch. Never execute it here:
  // the intended action is dragging it to the bookmarks bar.
  bookmarklet.addEventListener('click', function (event) {
    event.preventDefault();
    if (bookmarklet.getAttribute('aria-disabled') === 'true') {
      setStatus('The bookmarklet is still loading.');
      return;
    }
    setStatus('Drag this button to your bookmarks bar, then click it on your Twitch following page.');
  });
})();
