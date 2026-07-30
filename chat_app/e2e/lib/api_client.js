/**
 * API Client helper for E2E REST API interactions
 */

class ApiClient {
  constructor(baseUrl = process.env.BASE_URL || 'http://localhost:3000') {
    this.baseUrl = baseUrl.replace(/\/$/, '');
  }

  async request(method, path, body = null, token = null) {
    const url = `${this.baseUrl}${path.startsWith('/') ? path : '/' + path}`;
    const headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const options = {
      method,
      headers
    };

    if (body) {
      options.body = typeof body === 'string' ? body : JSON.stringify(body);
    }

    try {
      const response = await fetch(url, options);
      const status = response.status;
      let data = null;

      const contentType = response.headers.get('content-type') || '';
      if (contentType.includes('application/json')) {
        data = await response.json();
      } else {
        const text = await response.text();
        try {
          data = JSON.parse(text);
        } catch (_) {
          data = text;
        }
      }

      return {
        status,
        ok: response.ok,
        data,
        headers: response.headers
      };
    } catch (err) {
      return {
        status: 0,
        ok: false,
        error: err.message,
        data: null
      };
    }
  }

  // Auth endpoints
  async register(username, email, password) {
    return this.request('POST', '/api/auth/register', { username, email, password });
  }

  async login(email, password) {
    return this.request('POST', '/api/auth/login', { email, password });
  }

  async getMe(token) {
    return this.request('GET', '/api/auth/me', null, token);
  }

  // Friend endpoints
  async sendFriendRequest(token, target) {
    // target can be number (user_id), object, or string (username/email)
    let payload = {};
    if (typeof target === 'number') {
      payload = { to_user_id: target };
    } else if (typeof target === 'string') {
      if (target.includes('@')) {
        payload = { to_email: target };
      } else {
        payload = { to_username: target };
      }
    } else if (typeof target === 'object' && target !== null) {
      payload = target;
    }
    return this.request('POST', '/api/friends/request', payload, token);
  }

  async getPendingRequests(token) {
    return this.request('GET', '/api/friends/pending', null, token);
  }

  async acceptFriendRequest(token, requestId) {
    const payload = typeof requestId === 'object' ? requestId : { request_id: requestId };
    return this.request('POST', '/api/friends/accept', payload, token);
  }

  async getFriends(token) {
    return this.request('GET', '/api/friends', null, token);
  }

  // Message endpoints
  async getMessages(token, friendId) {
    return this.request('GET', `/api/messages/${friendId}`, null, token);
  }
}

module.exports = ApiClient;
