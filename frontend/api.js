/**
 * API Client for Recipe Backend
 * Handles all HTTP requests to the Go backend API
 */

'use strict';

// API Configuration
const API_CONFIG = {
  development: 'http://localhost:8080',
  production: 'https://your-app.fly.dev', // Update with your deployed API URL
};

// Get API base URL based on environment
function getAPIBaseURL() {
  // Check if running on GitHub Pages (production)
  const isProduction = window.location.hostname.includes('github.io');
  return isProduction ? API_CONFIG.production : API_CONFIG.development;
}

const API_BASE_URL = getAPIBaseURL();
const API_VERSION = '/api/v1';

// ── Error handling helper ──────────────────────────────────────────────────

class APIError extends Error {
  constructor(message, status, response) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.response = response;
  }
}

// ── HTTP request helper ────────────────────────────────────────────────────

async function request(endpoint, options = {}) {
  const url = `${API_BASE_URL}${API_VERSION}${endpoint}`;
  
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(url, config);

    // Handle 204 No Content
    if (response.status === 204) {
      return null;
    }

    // Parse JSON response
    const data = await response.json();

    // Handle non-2xx responses
    if (!response.ok) {
      throw new APIError(
        data.error || 'Request failed',
        response.status,
        data
      );
    }

    return data;
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    
    // Network or parsing error
    throw new APIError(
      `Network error: ${error.message}`,
      0,
      null
    );
  }
}

// ── Recipe API methods ─────────────────────────────────────────────────────

const RecipeAPI = {
  /**
   * Get all recipes with optional filtering
   * @param {Object} params - Query parameters
   * @param {string} params.category - Filter by category
   * @param {string} params.q - Search query
   * @returns {Promise<Array>} Array of recipes
   */
  async getAll(params = {}) {
    const queryParams = new URLSearchParams();
    if (params.category) queryParams.append('category', params.category);
    if (params.q) queryParams.append('q', params.q);
    
    const query = queryParams.toString();
    const endpoint = `/recipes${query ? '?' + query : ''}`;
    
    return request(endpoint);
  },

  /**
   * Get a single recipe by ID
   * @param {number} id - Recipe ID
   * @returns {Promise<Object>} Recipe object
   */
  async getById(id) {
    return request(`/recipes/${id}`);
  },

  /**
   * Create a new recipe
   * @param {Object} recipe - Recipe data
   * @returns {Promise<Object>} Created recipe with ID
   */
  async create(recipe) {
    return request('/recipes', {
      method: 'POST',
      body: JSON.stringify(recipe),
    });
  },

  /**
   * Update an existing recipe
   * @param {number} id - Recipe ID
   * @param {Object} recipe - Updated recipe data
   * @returns {Promise<Object>} Updated recipe
   */
  async update(id, recipe) {
    return request(`/recipes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(recipe),
    });
  },

  /**
   * Delete a recipe
   * @param {number} id - Recipe ID
   * @returns {Promise<null>} null on success
   */
  async delete(id) {
    return request(`/recipes/${id}`, {
      method: 'DELETE',
    });
  },
};

// ── Category API methods ───────────────────────────────────────────────────

const CategoryAPI = {
  /**
   * Get all unique categories
   * @returns {Promise<Array>} Array of category strings
   */
  async getAll() {
    return request('/categories');
  },
};

// ── Health check ───────────────────────────────────────────────────────────

const HealthAPI = {
  /**
   * Check if API is healthy
   * @returns {Promise<boolean>} true if healthy
   */
  async check() {
    try {
      const response = await fetch(`${API_BASE_URL}/health`);
      return response.ok;
    } catch {
      return false;
    }
  },
};

// ── Export API interface ───────────────────────────────────────────────────

const API = {
  recipes: RecipeAPI,
  categories: CategoryAPI,
  health: HealthAPI,
  baseURL: API_BASE_URL,
};

// Export for use in app.js
window.API = API;
