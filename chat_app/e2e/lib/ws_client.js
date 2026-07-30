/**
 * WebSocket Client helper for E2E real-time messaging tests
 */

const WebSocketImpl = globalThis.WebSocket || (() => {
  try { return require('ws'); } catch (_) { return null; }
})();

class WsClient {
  constructor(baseUrl = process.env.WS_URL || 'ws://localhost:3000') {
    this.baseUrl = baseUrl.replace(/\/$/, '');
    this.ws = null;
    this.receivedMessages = [];
    this.messageListeners = [];
    this.isOpen = false;
    this.connectionError = null;
  }

  connect(token, path = '/ws') {
    return new Promise((resolve, reject) => {
      if (!WebSocketImpl) {
        return reject(new Error('WebSocket implementation not available'));
      }

      const url = token ? `${this.baseUrl}${path}?token=${encodeURIComponent(token)}` : `${this.baseUrl}${path}`;
      
      try {
        this.ws = new WebSocketImpl(url);
      } catch (err) {
        return reject(err);
      }

      const timer = setTimeout(() => {
        if (!this.isOpen) {
          this.close();
          reject(new Error(`WebSocket connection timeout to ${url}`));
        }
      }, 5000);

      this.ws.onopen = () => {
        clearTimeout(timer);
        this.isOpen = true;
        resolve(this);
      };

      this.ws.onmessage = (event) => {
        let payload = event.data;
        if (typeof payload === 'string') {
          try { payload = JSON.parse(payload); } catch (_) {}
        }
        this.receivedMessages.push(payload);

        // Notify active listeners
        for (let i = this.messageListeners.length - 1; i >= 0; i--) {
          const listener = this.messageListeners[i];
          if (listener.predicate(payload)) {
            this.messageListeners.splice(i, 1);
            listener.resolve(payload);
          }
        }
      };

      this.ws.onerror = (err) => {
        this.connectionError = err;
        if (!this.isOpen) {
          clearTimeout(timer);
          reject(err);
        }
      };

      this.ws.onclose = () => {
        this.isOpen = false;
      };
    });
  }

  send(data) {
    if (!this.ws || !this.isOpen) {
      throw new Error('WebSocket is not connected');
    }
    const payload = typeof data === 'string' ? data : JSON.stringify(data);
    this.ws.send(payload);
  }

  waitForMessage(predicate = () => true, timeoutMs = 5000) {
    // Check if already received
    const existing = this.receivedMessages.find(predicate);
    if (existing) {
      return Promise.resolve(existing);
    }

    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        const idx = this.messageListeners.findIndex(l => l.resolve === resolve);
        if (idx !== -1) {
          this.messageListeners.splice(idx, 1);
        }
        reject(new Error(`Timeout waiting for WebSocket message after ${timeoutMs}ms`));
      }, timeoutMs);

      this.messageListeners.push({
        predicate,
        resolve: (msg) => {
          clearTimeout(timer);
          resolve(msg);
        }
      });
    });
  }

  close() {
    if (this.ws) {
      try {
        this.ws.close();
      } catch (_) {}
      this.isOpen = false;
    }
  }
}

module.exports = WsClient;
