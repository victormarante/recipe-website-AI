'use strict';

const SESSION_KEY = 'recipeWebsiteSession';

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

// ── Token helpers ──────────────────────────────────────────────────────────────

function getToken() {
  return sessionStorage.getItem(SESSION_KEY);
}

function setToken(token) {
  sessionStorage.setItem(SESSION_KEY, token);
}

function clearToken() {
  sessionStorage.removeItem(SESSION_KEY);
}

// ── Error class ────────────────────────────────────────────────────────────────

class APIError extends Error {
  constructor(message, status, response) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.response = response;
  }
}

// ── HTTP request helper ────────────────────────────────────────────────────────

async function request(endpoint, options = {}) {
  const url = `${API_BASE_URL}${API_VERSION}${endpoint}`;
  const token = getToken();

  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(url, config);

    // Token expired or invalid — force re-login
    if (response.status === 401) {
      clearToken();
      location.reload();
      return;
    }

    if (response.status === 204) {
      return null;
    }

    const data = await response.json();

    if (!response.ok) {
      throw new APIError(data.error || 'Request failed', response.status, data);
    }

    return data;
  } catch (error) {
    if (error instanceof APIError) throw error;
    throw new APIError(`Network error: ${error.message}`, 0, null);
  }
}

// ── Auth API ───────────────────────────────────────────────────────────────────

const AuthAPI = {
  async login(username, password) {
    const url = `${API_BASE_URL}${API_VERSION}/auth/login`;
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (!response.ok) {
      throw new APIError('Invalid credentials', response.status, null);
    }
    return response.json();
  },
};

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
};

// ── Category API ───────────────────────────────────────────────────────────────

const CategoryAPI = {
  async getAll() {
    return request('/categories');
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
  auth: AuthAPI,
  recipes: RecipeAPI,
  categories: CategoryAPI,
  health: HealthAPI,
  baseURL: API_BASE_URL,
  getToken,
  setToken,
  clearToken,
};
