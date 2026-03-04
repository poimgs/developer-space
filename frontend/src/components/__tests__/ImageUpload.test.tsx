import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import ImageUpload from '../ImageUpload';

// --- Mocks ---

const mockUploadSessionImage = vi.fn();
const mockDelete = vi.fn();
const mockAddToast = vi.fn();

vi.mock('../../api/client', () => ({
  api: {
    uploadSessionImage: (...args: unknown[]) => mockUploadSessionImage(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, body: { error: string }) {
      super(body.error);
      this.status = status;
    }
  },
}));

vi.mock('../../context/ToastContext', () => ({
  useToast: () => ({ addToast: mockAddToast }),
}));

// --- Helpers ---

function makeFile(name: string, type: string, sizeBytes = 1024): File {
  const buffer = new ArrayBuffer(sizeBytes);
  return new File([buffer], name, { type });
}

function renderUpload(props: {
  sessionId?: string;
  currentImageUrl?: string | null;
  onUpload?: () => void;
  onRemove?: () => void;
} = {}) {
  const defaultProps = {
    sessionId: 'session-1',
    currentImageUrl: null as string | null,
    onUpload: vi.fn(),
    onRemove: vi.fn(),
    ...props,
  };
  return { ...render(<ImageUpload {...defaultProps} />), props: defaultProps };
}

describe('ImageUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // --- Empty state (no image) ---

  it('renders drop zone when no image exists', () => {
    renderUpload();
    expect(screen.getByText('Click or drag to upload')).toBeInTheDocument();
    expect(screen.getByText('JPEG, PNG, or WebP, max 5MB')).toBeInTheDocument();
  });

  it('has label associated with file input', () => {
    renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    expect(input).toHaveAttribute('id', 'image-upload-session-session-1');
  });

  it('renders visually-hidden file input with correct accept attribute', () => {
    renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    expect(input).toHaveAttribute('accept', 'image/jpeg,image/png,image/webp,image/*');
    expect(input).toHaveClass('sr-only');
  });

  // --- Existing image state ---

  it('renders image preview when currentImageUrl is set', () => {
    renderUpload({ currentImageUrl: '/uploads/sessions/test.jpg' });
    const img = screen.getByAltText('Session');
    expect(img).toHaveAttribute('src', '/uploads/sessions/test.jpg');
  });

  it('shows Replace and Remove buttons when image exists', () => {
    renderUpload({ currentImageUrl: '/uploads/sessions/test.jpg' });
    expect(screen.getByText('Replace')).toBeInTheDocument();
    expect(screen.getByText('Remove')).toBeInTheDocument();
  });

  // --- Client-side validation ---

  it('shows error toast for invalid file type', async () => {
    renderUpload();

    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('doc.pdf', 'application/pdf');
    // Use fireEvent to bypass HTML accept attribute filtering
    fireEvent.change(input, { target: { files: [file] } });

    await waitFor(() => {
      expect(mockAddToast).toHaveBeenCalledWith(
        'Invalid file type. Please upload JPEG, PNG, or WebP.',
        'error',
      );
    });
    expect(mockUploadSessionImage).not.toHaveBeenCalled();
  });

  it('shows error toast for file exceeding 5MB', async () => {
    const user = userEvent.setup();
    renderUpload();

    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('big.jpg', 'image/jpeg', 6 * 1024 * 1024);
    await user.upload(input, file);

    expect(mockAddToast).toHaveBeenCalledWith(
      'File too large. Maximum size is 5MB.',
      'error',
    );
    expect(mockUploadSessionImage).not.toHaveBeenCalled();
  });

  // --- Successful upload ---

  it('uploads valid JPEG and calls onUpload with returned URL', async () => {
    const user = userEvent.setup();
    mockUploadSessionImage.mockResolvedValue({
      data: { image_url: '/uploads/sessions/session-1-123.jpg' },
    });

    const { props } = renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('photo.jpg', 'image/jpeg');
    await user.upload(input, file);

    await waitFor(() => {
      expect(mockUploadSessionImage).toHaveBeenCalledWith('session-1', file);
    });
    expect(props.onUpload).toHaveBeenCalledWith('/uploads/sessions/session-1-123.jpg');
    expect(mockAddToast).toHaveBeenCalledWith('Image uploaded.', 'success');
  });

  it('uploads valid PNG file', async () => {
    const user = userEvent.setup();
    mockUploadSessionImage.mockResolvedValue({
      data: { image_url: '/uploads/sessions/session-1-456.png' },
    });

    const { props } = renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('image.png', 'image/png');
    await user.upload(input, file);

    await waitFor(() => {
      expect(props.onUpload).toHaveBeenCalledWith('/uploads/sessions/session-1-456.png');
    });
  });

  it('uploads valid WebP file', async () => {
    const user = userEvent.setup();
    mockUploadSessionImage.mockResolvedValue({
      data: { image_url: '/uploads/sessions/session-1-789.webp' },
    });

    const { props } = renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('image.webp', 'image/webp');
    await user.upload(input, file);

    await waitFor(() => {
      expect(props.onUpload).toHaveBeenCalledWith('/uploads/sessions/session-1-789.webp');
    });
  });

  // --- Upload error ---

  it('shows error toast when upload API fails', async () => {
    const user = userEvent.setup();
    const { ApiError } = await import('../../api/client');
    mockUploadSessionImage.mockRejectedValue(
      new ApiError(422, { error: 'Unsupported file type' }),
    );

    const { props } = renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('photo.jpg', 'image/jpeg');
    await user.upload(input, file);

    await waitFor(() => {
      expect(mockAddToast).toHaveBeenCalledWith('Unsupported file type', 'error');
    });
    expect(props.onUpload).not.toHaveBeenCalled();
  });

  it('shows generic error for non-ApiError upload failure', async () => {
    const user = userEvent.setup();
    mockUploadSessionImage.mockRejectedValue(new Error('Network error'));

    renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('photo.jpg', 'image/jpeg');
    await user.upload(input, file);

    await waitFor(() => {
      expect(mockAddToast).toHaveBeenCalledWith('Failed to upload image', 'error');
    });
  });

  // --- Remove image ---

  it('calls delete API and onRemove when Remove is clicked', async () => {
    const user = userEvent.setup();
    mockDelete.mockResolvedValue(undefined);

    const { props } = renderUpload({ currentImageUrl: '/uploads/sessions/test.jpg' });
    await user.click(screen.getByText('Remove'));

    await waitFor(() => {
      expect(mockDelete).toHaveBeenCalledWith('/api/sessions/session-1/image');
    });
    expect(props.onRemove).toHaveBeenCalled();
    expect(mockAddToast).toHaveBeenCalledWith('Image removed.', 'success');
  });

  it('shows error toast when remove API fails', async () => {
    const user = userEvent.setup();
    const { ApiError } = await import('../../api/client');
    mockDelete.mockRejectedValue(
      new ApiError(404, { error: 'Session not found' }),
    );

    const { props } = renderUpload({ currentImageUrl: '/uploads/sessions/test.jpg' });
    await user.click(screen.getByText('Remove'));

    await waitFor(() => {
      expect(mockAddToast).toHaveBeenCalledWith('Session not found', 'error');
    });
    expect(props.onRemove).not.toHaveBeenCalled();
  });

  // --- Loading state ---

  it('shows uploading spinner during upload', async () => {
    // Never-resolving promise to keep uploading state
    mockUploadSessionImage.mockReturnValue(new Promise(() => {}));

    const user = userEvent.setup();
    renderUpload();
    const input = screen.getByLabelText('Upload session image', { selector: 'input' });
    const file = makeFile('photo.jpg', 'image/jpeg');
    await user.upload(input, file);

    expect(screen.getByText('Uploading...')).toBeInTheDocument();
  });

  it('disables Replace and Remove during removal', async () => {
    mockDelete.mockReturnValue(new Promise(() => {}));

    const user = userEvent.setup();
    renderUpload({ currentImageUrl: '/uploads/sessions/test.jpg' });
    await user.click(screen.getByText('Remove'));

    expect(screen.getByText('Replace')).toHaveClass('pointer-events-none');
    expect(screen.getByText('Remove')).toBeDisabled();
  });
});
