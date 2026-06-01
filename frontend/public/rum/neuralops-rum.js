/**
 * NeuralOps RUM SDK — RUM beacons + rrweb session replay (when rrweb is loaded).
 *
 * Usage:
 *   <script src="https://cdn.jsdelivr.net/npm/rrweb@latest/dist/rrweb.min.js"></script>
 *   <script src="/rum/neuralops-rum.js"></script>
 *   NeuralOpsRUM.init({ endpoint: '/api/v1', tenantId: '...', replayEnabled: true });
 */
(function (global) {
  'use strict';

  var config = {
    endpoint: '/api/v1',
    apiKey: '',
    tenantId: '',
    sessionId: '',
    userId: '',
    replayEnabled: true,
    flushIntervalMs: 10000,
    maskInputs: true,
  };

  var seq = 0;
  var startTime = Date.now();
  var stopRecord = null;

  function headers() {
    var h = { 'Content-Type': 'application/json' };
    if (config.apiKey) h['X-API-Key'] = config.apiKey;
    if (config.tenantId) h['X-Tenant-ID'] = config.tenantId;
    return h;
  }

  function post(path, body) {
    return fetch(config.endpoint + path, {
      method: 'POST',
      headers: headers(),
      body: JSON.stringify(body),
      keepalive: true,
    }).catch(function () { /* swallow */ });
  }

  function sessionId() {
    if (!config.sessionId) {
      config.sessionId = 'rum-' + Math.random().toString(36).slice(2, 12) + '-' + Date.now();
    }
    return config.sessionId;
  }

  function hasConsent() {
    try {
      return localStorage.getItem('neuralops_rum_consent') === 'granted';
    } catch (_) {
      return false;
    }
  }

  function beacon() {
    var nav = performance.getEntriesByType('navigation')[0];
    post('/rum/beacon', {
      sessionId: sessionId(),
      userId: config.userId,
      page: location.pathname,
      device: /Mobi/i.test(navigator.userAgent) ? 'mobile' : 'desktop',
      country: 'unknown',
      durationMs: Date.now() - startTime,
      errors: window.__neuralopsRumErrors || 0,
      lcp: nav ? nav.domContentLoadedEventEnd : 0,
    });
  }

  function recordReplay(type, payload) {
    if (!config.replayEnabled || !hasConsent()) return;
    seq += 1;
    post('/rum/replay', {
      sessionId: sessionId(),
      seq: seq,
      type: type,
      payload: payload || {},
    });
  }

  function startRrweb() {
    if (!global.rrweb || !config.replayEnabled || !hasConsent()) return;
    stopRecord = global.rrweb.record({
      emit: function (event) {
        recordReplay('rrweb', { event: event });
      },
      maskAllInputs: config.maskInputs,
      maskTextSelector: '[data-private]',
    });
  }

  function init(opts) {
    Object.assign(config, opts || {});
    window.__neuralopsRumErrors = 0;
    window.addEventListener('error', function () {
      window.__neuralopsRumErrors += 1;
    });

    if (hasConsent()) {
      startRrweb();
      if (!global.rrweb) {
        document.addEventListener('click', function (e) {
          var t = e.target;
          recordReplay('click', { x: e.clientX, y: e.clientY, tag: t && t.tagName });
        });
      }
    }

    window.addEventListener('neuralops-rum-consent', function (ev) {
      if (ev.detail && ev.detail.granted) startRrweb();
    });

    beacon();
    setInterval(beacon, config.flushIntervalMs);
  }

  function identify(userId) {
    config.userId = userId;
  }

  function stop() {
    if (typeof stopRecord === 'function') stopRecord();
  }

  global.NeuralOpsRUM = { init: init, identify: identify, record: recordReplay, stop: stop };
})(typeof window !== 'undefined' ? window : globalThis);
