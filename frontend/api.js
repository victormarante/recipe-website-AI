'use strict';

const API_CONFIG = {
  development: 'http://localhost:8080',
  production: 'https://recipe-website-ai.fly.dev',
};

function getAPIBaseURL() {
  const isProduction = window.location.hostname.includes('github.io');
  return isProduction ? API_CONFIG.production : API_CONFIG.development;
}

const API_BASE_URL = getAPIBaseURL();
const API_VERSION = '/api/v1';

// ── Error class ────────────────────────────────────────────────────────────────

class APIError extends Error {
  constructor(message, status, response) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.response = response;
  }
}

// ── Auth token ─────────────────────────────────────────────────────────────────

const TOKEN_KEY = 'recipe_admin_token';

function getToken() {
  try { return localStorage.getItem(TOKEN_KEY); } catch { return null; }
}

function setToken(token) {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token);
    else localStorage.removeItem(TOKEN_KEY);
  } catch { /* storage unavailable; token only lives for this page */ }
  window.dispatchEvent(new Event('auth-changed'));
}

function authHeaders() {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// A 401 on a request that sent a token means it expired or is invalid.
function handleUnauthorized() {
  if (getToken()) setToken(null);
}

// ── HTTP request helper ────────────────────────────────────────────────────────

async function request(endpoint, options = {}) {
  const url = `${API_BASE_URL}${API_VERSION}${endpoint}`;

  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(url, config);

    if (response.status === 204) {
      return null;
    }

    // Non-JSON bodies (proxy errors, old deploys, plain-text 404s) must not leak parser errors.
    const data = await response.json().catch(() => null);

    if (!response.ok) {
      if (response.status === 401 && endpoint !== '/auth/login') handleUnauthorized();
      throw new APIError((data && data.error) || t('Request failed'), response.status, data);
    }

    if (data === null) {
      throw new APIError(t('Unexpected server response'), response.status, null);
    }

    return data;
  } catch (error) {
    if (error instanceof APIError) throw error;
    throw new APIError(t('Network error'), 0, null);
  }
}

// ── Recipe API ─────────────────────────────────────────────────────────────────

const RecipeAPI = {
  async getAll(params = {}) {
    const queryParams = new URLSearchParams();
    if (params.category) queryParams.append('category', params.category);
    if (params.q) queryParams.append('q', params.q);
    const query = queryParams.toString();
    return request(`/recipes${query ? '?' + query : ''}`);
  },

  async getById(id) {
    return request(`/recipes/${id}`);
  },

  async create(recipe) {
    return request('/recipes', {
      method: 'POST',
      body: JSON.stringify(recipe),
    });
  },

  async update(id, recipe) {
    return request(`/recipes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(recipe),
    });
  },

  async delete(id) {
    return request(`/recipes/${id}`, { method: 'DELETE' });
  },

  async uploadImage(id, file) {
    const url = `${API_BASE_URL}${API_VERSION}/recipes/${id}/image`;
    const form = new FormData();
    form.append('image', file);
    const response = await fetch(url, {
      method: 'POST',
      headers: authHeaders(),
      body: form,
    });
    if (!response.ok) {
      if (response.status === 401) handleUnauthorized();
      const data = await response.json().catch(() => ({}));
      throw new APIError(data.error || t('Upload failed'), response.status, data);
    }
    return response.json();
  },

  async deleteImage(id) {
    return request(`/recipes/${id}/image`, { method: 'DELETE' });
  },
};

// ── Category API ───────────────────────────────────────────────────────────────

const CategoryAPI = {
  async getAll() {
    return request('/categories');
  },
};

// ── Auth API ───────────────────────────────────────────────────────────────────

const AuthAPI = {
  async login(pin) {
    const data = await request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ pin }),
    });
    setToken(data.token);
    return data;
  },

  logout() {
    setToken(null);
  },

  isLoggedIn() {
    return Boolean(getToken());
  },
};

// ── Health check ───────────────────────────────────────────────────────────────

const HealthAPI = {
  async check() {
    try {
      const response = await fetch(`${API_BASE_URL}/health`);
      return response.ok;
    } catch {
      return false;
    }
  },
};

// ── Exports ────────────────────────────────────────────────────────────────────

window.API = {
  recipes: RecipeAPI,
  categories: CategoryAPI,
  auth: AuthAPI,
  health: HealthAPI,
  baseURL: API_BASE_URL,
};
