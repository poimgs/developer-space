import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from '../client';

describe('API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    globalThis.fetch = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  function mockFetch(status: number, body: unknown, ok = status >= 200 && status < 300) {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok,
      status,
      statusText: 'Not Found',
      json: () => Promise.resolve(body),
    });
  }

  it('GET sends credentials and Content-Type', async () => {
    mockFetch(200, { data: 'ok' });
    await api.get('/api/test');
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      credentials: 'include',
      headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
    }));
  });

  it('GET returns parsed JSON', async () => {
    mockFetch(200, { data: { id: '123' } });
    const result = await api.get('/api/test');
    expect(result).toEqual({ data: { id: '123' } });
  });

  it('POST sends method and body', async () => {
    mockFetch(201, { data: { id: '123' } });
    await api.post('/api/test', { name: 'foo' });
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ name: 'foo' }),
    }));
  });

  it('PATCH sends method and body', async () => {
    mockFetch(200, { data: { id: '123' } });
    await api.patch('/api/test/123', { name: 'bar' });
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test/123', expect.objectContaining({
      method: 'PATCH',
      body: JSON.stringify({ name: 'bar' }),
    }));
  });

  it('DELETE sends method', async () => {
    mockFetch(204, undefined);
    await api.delete('/api/test/123');
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test/123', expect.objectContaining({
      method: 'DELETE',
    }));
  });

  it('returns undefined for 204 responses', async () => {
    mockFetch(204, undefined);
    const result = await api.delete('/api/test/123');
    expect(result).toBeUndefined();
  });

  it('throws ApiError on non-ok response', async () => {
    mockFetch(404, { error: 'not found' }, false);
    await expect(api.get('/api/test')).rejects.toThrow(ApiError);
  });

  it('ApiError contains status and body', async () => {
    mockFetch(422, { error: 'validation failed', details: { name: 'required' } }, false);
    try {
      await api.post('/api/test', {});
      expect.fail('should throw');
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      const apiErr = err as ApiError;
      expect(apiErr.status).toBe(422);
      expect(apiErr.body.error).toBe('validation failed');
      expect(apiErr.body.details).toEqual({ name: 'required' });
      expect(apiErr.message).toBe('validation failed');
    }
  });

  it('handles non-JSON error response', async () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      json: () => Promise.reject(new Error('not json')),
    });
    try {
      await api.get('/api/test');
      expect.fail('should throw');
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).message).toBe('Internal Server Error');
    }
  });

  describe('getPublicProfile', () => {
    it('fetches profile by ID', async () => {
      const profile = {
        id: 'abc-123',
        name: 'Jane Doe',
        telegram_handle: 'janedoe',
        bio: 'Developer',
        skills: ['go', 'react'],
        linkedin_url: 'https://linkedin.com/in/jane',
        instagram_handle: 'jane',
        github_username: 'janedoe',
      };
      mockFetch(200, { data: profile });

      const result = await api.getPublicProfile('abc-123');
      expect(result).toEqual({ data: profile });
      expect(globalThis.fetch).toHaveBeenCalledWith(
        '/api/profiles/abc-123',
        expect.objectContaining({
          credentials: 'include',
          headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
        }),
      );
    });

    it('throws ApiError when profile not found', async () => {
      mockFetch(404, { error: 'not found' }, false);
      await expect(api.getPublicProfile('nonexistent')).rejects.toThrow(ApiError);
    });
  });

  describe('uploadSessionImage', () => {
    it('sends multipart form with image file', async () => {
      const responseBody = { data: { image_url: '/uploads/sessions/abc-123-1234567890.jpg' } };
      (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(responseBody),
      });

      const file = new File(['fake-image-data'], 'photo.jpg', { type: 'image/jpeg' });
      const result = await api.uploadSessionImage('abc-123', file);

      expect(result).toEqual(responseBody);
      expect(globalThis.fetch).toHaveBeenCalledWith(
        '/api/sessions/abc-123/image',
        expect.objectContaining({
          method: 'POST',
          credentials: 'include',
        }),
      );

      // Verify FormData was sent (not JSON Content-Type)
      const callArgs = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      expect(callArgs[1].body).toBeInstanceOf(FormData);
      expect(callArgs[1].headers).toBeUndefined();
    });

    it('throws ApiError on upload failure', async () => {
      (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: false,
        status: 413,
        statusText: 'Payload Too Large',
        json: () => Promise.resolve({ error: 'file too large' }),
      });

      const file = new File(['data'], 'big.jpg', { type: 'image/jpeg' });
      await expect(api.uploadSessionImage('abc-123', file)).rejects.toThrow(ApiError);
    });
  });
});
