// src/api.ts
import axios from 'axios';
import { getAccessToken, setAccessToken } from './auth';
import { API_BASE_URL } from './config';

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,      // ensure refresh cookie is sent
});

// attach access token on each request
api.interceptors.request.use(config => {
  const token = getAccessToken();
  if (token) config.headers!['Authorization'] = `Bearer ${token}`;
  return config;
});

// on 401, try to refresh once and retry original
api.interceptors.response.use(
  res => res,
  async err => {
    if (err.response?.status === 401 && !err.config.__isRetry) {
      err.config.__isRetry = true;
      const { data } = await axios.post(
        `${API_BASE_URL}/user/refresh`,
        {},
        { withCredentials: true }
      );
      setAccessToken(data.access_token);
      err.config.headers['Authorization'] = `Bearer ${data.access_token}`;
      return axios(err.config);
    }
    return Promise.reject(err);
  }
);

export default api;
