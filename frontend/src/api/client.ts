import type { APIError, APIResponse, PublicMember } from '../types';

export class ApiError extends Error {
  status: number;
  body: APIError;

  constructor(status: number, body: APIError) {
    super(body.error);
    this.status = status;
    this.body = body;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

async function uploadRequest<T>(path: string, file: File): Promise<T> {
  const form = new FormData();
  form.append('image', file);

  const res = await fetch(path, {
    method: 'POST',
    credentials: 'include',
    body: form,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body);
  }

  return res.json();
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),

  getPublicProfile: (id: string) =>
    request<APIResponse<PublicMember>>(`/api/profiles/${id}`),

  getSkills: () =>
    request<APIResponse<string[]>>('/api/skills'),

  uploadSessionImage: (sessionId: string, file: File) =>
    uploadRequest<APIResponse<{ image_url: string }>>(`/api/sessions/${sessionId}/image`, file),

  updateSeries: (seriesId: string, data: unknown) =>
    request<APIResponse<unknown>>(`/api/sessions/series/${seriesId}`, { method: 'PATCH', body: JSON.stringify(data) }),

  cancelSeries: (seriesId: string) =>
    request<unknown>(`/api/sessions/series/${seriesId}`, { method: 'DELETE' }),

  uploadSeriesImage: (seriesId: string, file: File) =>
    uploadRequest<APIResponse<{ image_url: string }>>(`/api/sessions/series/${seriesId}/image`, file),

  deleteSeriesImage: (seriesId: string) =>
    request<unknown>(`/api/sessions/series/${seriesId}/image`, { method: 'DELETE' }),
};
