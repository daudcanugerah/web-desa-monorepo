/**
 * API Client Configuration
 * Centralized fetch-based HTTP client for all API calls
 */

const RAW_API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';

// Ensure the base URL always carries the `/api/v1` prefix. Accepts either
// `http://host:port` or `http://host:port/api/v1` from VITE_API_BASE_URL.
const normalizeBaseUrl = (base) => {
  const trimmed = String(base).replace(/\/+$/, '');
  return /\/api\/v\d+$/.test(trimmed) ? trimmed : `${trimmed}/api/v1`;
};

const API_BASE_URL = normalizeBaseUrl(RAW_API_BASE_URL);

const TOKEN_KEY = 'auth_token';

class APIClient {
  constructor(baseURL) {
    this.baseURL = baseURL;
  }

  /**
   * Get stored auth token
   */
  getToken() {
    return localStorage.getItem(TOKEN_KEY);
  }

  /**
   * Set auth token
   */
  setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
  }

  /**
   * Clear auth token
   */
  clearToken() {
    localStorage.removeItem(TOKEN_KEY);
  }

  /**
   * Create headers object with optional auth
   */
  getHeaders(options = {}) {
    const headers = {
      ...options.headers,
    };

    if (!options.skipAuth && this.getToken()) {
      headers['Authorization'] = `Bearer ${this.getToken()}`;
    }

    return headers;
  }

  /**
   * Create multipart form data for file uploads
   */
  createFormData(data) {
    const formData = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((item, index) => {
          formData.append(`${key}[${index}]`, item);
        });
      } else if (value !== undefined && value !== null) {
        formData.append(key, value);
      }
    });
    return formData;
  }

  /**
   * Make HTTP request
   * @param {string} method - HTTP method (GET, POST, etc.)
   * @param {string} endpoint - API endpoint
   * @param {object} options - Request options (data, headers, params, isMultipart)
   * @returns {Promise<object>} Response data
   */
  async request(method, endpoint, options = {}) {
    const url = new URL(`${this.baseURL}${endpoint}`);

    if (options.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, value);
        }
      });
    }

    const isMultipart = options.isMultipart || false;
    const headers = this.getHeaders(options);

    const config = {
      method,
      headers,
    };

    if (options.data && method !== 'GET') {
      if (isMultipart) {
        config.body = this.createFormData(options.data);
        delete config.headers['Content-Type'];
      } else {
        config.body = JSON.stringify(options.data);
      }
    }

    try {
      const response = await fetch(url.toString(), config);

      const contentType = response.headers.get('content-type');
      let data;

      if (contentType && contentType.includes('application/json')) {
        data = await response.json();
      } else {
        data = await response.text();
      }

      if (!response.ok) {
        const error = new Error(data?.error || `HTTP ${response.status}`);
        error.status = response.status;
        error.data = data;
        throw error;
      }

      return data;
    } catch (error) {
      console.error(`API Error [${method} ${endpoint}]:`, error);
      throw error;
    }
  }

  get(endpoint, options = {}) {
    return this.request('GET', endpoint, options);
  }

  post(endpoint, data, options = {}) {
    return this.request('POST', endpoint, { ...options, data });
  }

  put(endpoint, data, options = {}) {
    return this.request('PUT', endpoint, { ...options, data });
  }

  patch(endpoint, data, options = {}) {
    return this.request('PATCH', endpoint, { ...options, data });
  }

  delete(endpoint, options = {}) {
    return this.request('DELETE', endpoint, options);
  }
}

export const apiClient = new APIClient(API_BASE_URL);

export { API_BASE_URL };

export default apiClient;
