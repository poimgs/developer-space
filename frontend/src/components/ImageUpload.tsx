import { useRef, useState, type DragEvent } from 'react';
import { api, ApiError } from '../api/client';
import { useToast } from '../context/ToastContext';

interface ImageUploadProps {
  sessionId: string;
  currentImageUrl?: string | null;
  onUpload: (imageUrl: string) => void;
  onRemove?: () => void;
}

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp'];
const MAX_SIZE = 5 * 1024 * 1024; // 5MB

export default function ImageUpload({ sessionId, currentImageUrl, onUpload, onRemove }: ImageUploadProps) {
  const { addToast } = useToast();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);

  function validateFile(file: File): string | null {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return 'Invalid file type. Please upload JPEG, PNG, or WebP.';
    }
    if (file.size > MAX_SIZE) {
      return 'File too large. Maximum size is 5MB.';
    }
    return null;
  }

  async function handleFile(file: File) {
    const error = validateFile(file);
    if (error) {
      addToast(error, 'error');
      return;
    }

    setUploading(true);
    try {
      const res = await api.uploadSessionImage(sessionId, file);
      onUpload(res.data.image_url);
      addToast('Image uploaded.', 'success');
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Failed to upload image';
      addToast(message, 'error');
    } finally {
      setUploading(false);
    }
  }

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
    if (fileInputRef.current) fileInputRef.current.value = '';
  }

  function handleDrop(e: DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) handleFile(file);
  }

  function handleDragOver(e: DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setDragOver(true);
  }

  function handleDragLeave(e: DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setDragOver(false);
  }

  async function handleRemove() {
    setUploading(true);
    try {
      await api.delete(`/api/sessions/${sessionId}/image`);
      onRemove?.();
      addToast('Image removed.', 'success');
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Failed to remove image';
      addToast(message, 'error');
    } finally {
      setUploading(false);
    }
  }

  if (currentImageUrl) {
    return (
      <div className="relative rounded-xl overflow-hidden">
        <img
          src={currentImageUrl}
          alt="Session"
          className="w-full h-40 object-cover"
        />
        <div className="absolute inset-0 bg-black/0 hover:bg-black/40 transition-colors flex items-center justify-center opacity-0 hover:opacity-100">
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            className="rounded-md bg-white px-3 py-1.5 text-sm font-medium text-stone-700 mr-2 disabled:opacity-50"
          >
            Replace
          </button>
          <button
            onClick={handleRemove}
            disabled={uploading}
            className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            Remove
          </button>
        </div>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          onChange={handleInputChange}
          className="hidden"
          aria-label="Upload session image"
        />
      </div>
    );
  }

  return (
    <div>
      <div
        onClick={() => !uploading && fileInputRef.current?.click()}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={`border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors ${
          dragOver
            ? 'border-amber-500 bg-amber-50 dark:border-amber-400 dark:bg-amber-900/20'
            : 'border-stone-300 hover:border-amber-500 dark:border-stone-600 dark:hover:border-amber-400'
        } ${uploading ? 'opacity-50 pointer-events-none' : ''}`}
        role="button"
        aria-label="Upload session image"
      >
        {uploading ? (
          <div className="flex flex-col items-center">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-amber-600 border-t-transparent" />
            <p className="mt-2 text-sm text-stone-500 dark:text-stone-400">Uploading...</p>
          </div>
        ) : (
          <>
            <svg className="mx-auto h-10 w-10 text-stone-400 dark:text-stone-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
            </svg>
            <p className="mt-2 text-sm text-stone-600 dark:text-stone-400">Click or drag to upload</p>
            <p className="mt-1 text-xs text-stone-400 dark:text-stone-500">JPEG, PNG, or WebP, max 5MB</p>
          </>
        )}
      </div>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        onChange={handleInputChange}
        className="hidden"
        aria-label="Upload session image"
      />
    </div>
  );
}
