export const ENDPOINTS = {
  // Auth
  LOGIN: '/auth/login',
  REGISTER: '/auth/register',
  REFRESH: '/auth/refresh',
  LOGOUT: '/auth/logout',
  PROFILE: '/auth/me',
  
  // Users
  USERS: '/users',
  USER_BY_ID: (id: string | number) => `/users/${id}`,
  
  // Products
  PRODUCTS: '/products',
  PRODUCT_BY_ID: (id: string) => `/products/${id}`,
  MY_PRODUCTS: '/products/me',
  
  // Chat endpoints - Conformes à la documentation backend
  CHAT_CONVERSATIONS: '/api/v1/chat/conversations',
  CHAT_ROOMS: '/api/v1/chat/rooms',
  CHAT_ROOM_MESSAGES: (roomName: string) => `/api/v1/chat/rooms/${roomName}/messages`,
  CHAT_DM: (userId: string | number) => `/api/v1/chat/dm/${userId}`,
  CHAT_DM_READ: (userId: string | number) => `/api/v1/chat/messages/${userId}/read`,
  CHAT_UNREAD: '/api/v1/chat/unread',
  CHAT_CREATE_ROOM: '/api/v1/chat/rooms',
  CHAT_USERS: '/api/v1/users/except-me', // Utilisateurs pour les DM
  
  // Tracks
  TRACKS: '/tracks',
  TRACK_BY_ID: (id: string | number) => `/tracks/${id}`,
  TRACK_UPLOAD: '/tracks/upload',
  
  // Resources
  RESOURCES: '/resources',
  RESOURCE_BY_ID: (id: string | number) => `/resources/${id}`,
  RESOURCE_UPLOAD: '/resources/upload',
  
  // Admin
  ADMIN_STATS: '/admin/stats',
  ADMIN_USERS: '/admin/users',
  ADMIN_USER_BY_ID: (id: string | number) => `/admin/users/${id}`,
  ADMIN_TRACKS: '/admin/tracks',
  ADMIN_TRACK_BY_ID: (id: string | number) => `/admin/tracks/${id}`,
  ADMIN_RESOURCES: '/admin/resources',
  ADMIN_RESOURCE_BY_ID: (id: string | number) => `/admin/resources/${id}`,
} as const; 