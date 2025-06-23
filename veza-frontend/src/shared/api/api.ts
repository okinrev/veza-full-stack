import axios from "axios";
import { cacheService } from "@/lib/cache";

export const api = axios.create({
  baseURL: "http://localhost:8080/api/v1",
  headers: {
    "Content-Type": "application/json",
    "Accept": "application/json",
  },
  withCredentials: false
});

// Intercepteur pour le cache
api.interceptors.request.use(async (config) => {
  if (config.method?.toLowerCase() === 'get') {
    const cacheKey = `${config.url}${JSON.stringify(config.params || {})}`;
    const cachedData = cacheService.get(cacheKey);
    
    if (cachedData) {
      return Promise.reject({
        __CACHE__: true,
        data: cachedData
      });
    }
  }
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Intercepteur pour gérer les erreurs
api.interceptors.response.use(
  (response) => {
    if (response.config.method?.toLowerCase() === 'get') {
      const cacheKey = `${response.config.url}${JSON.stringify(response.config.params || {})}`;
      cacheService.set(cacheKey, response.data);
    }
    return response;
  },
  (error) => {
    if (error.__CACHE__) {
      return Promise.resolve({ data: error.data });
    }
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
); 