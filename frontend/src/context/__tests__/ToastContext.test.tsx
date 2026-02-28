import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Toast from '../../components/Toast';
import { ToastProvider, useToast } from '../ToastContext';

function TestTrigger() {
  const { addToast } = useToast();
  return (
    <div>
      <button onClick={() => addToast('Success!', 'success')}>Add Success</button>
      <button onClick={() => addToast('Error!', 'error')}>Add Error</button>
      <button onClick={() => addToast('Info!')}>Add Info</button>
    </div>
  );
}

function renderWithToast() {
  return render(
    <ToastProvider>
      <Toast />
      <TestTrigger />
    </ToastProvider>,
  );
}

describe('ToastContext', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows a toast when addToast is called', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Success'));
    expect(screen.getByText('Success!')).toBeInTheDocument();
  });

  it('shows multiple toasts', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Success'));
    fireEvent.click(screen.getByText('Add Error'));
    expect(screen.getByText('Success!')).toBeInTheDocument();
    expect(screen.getByText('Error!')).toBeInTheDocument();
  });

  it('limits to 3 visible toasts', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Success'));
    fireEvent.click(screen.getByText('Add Error'));
    fireEvent.click(screen.getByText('Add Info'));
    fireEvent.click(screen.getByText('Add Success'));
    const toasts = screen.getAllByRole('button', { name: 'Dismiss' });
    expect(toasts.length).toBeLessThanOrEqual(3);
  });

  it('dismisses toast on X click', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Success'));
    expect(screen.getByText('Success!')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }));
    expect(screen.queryByText('Success!')).toBeNull();
  });

  it('auto-dismisses success toast after 3s', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Success'));
    expect(screen.getByText('Success!')).toBeInTheDocument();
    act(() => { vi.advanceTimersByTime(3000); });
    expect(screen.queryByText('Success!')).toBeNull();
  });

  it('auto-dismisses error toast after 5s', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Error'));
    expect(screen.getByText('Error!')).toBeInTheDocument();
    act(() => { vi.advanceTimersByTime(3000); });
    expect(screen.getByText('Error!')).toBeInTheDocument(); // Still there at 3s
    act(() => { vi.advanceTimersByTime(2000); });
    expect(screen.queryByText('Error!')).toBeNull(); // Gone at 5s
  });

  it('defaults to info type with indigo border', () => {
    renderWithToast();
    fireEvent.click(screen.getByText('Add Info'));
    const toast = screen.getByText('Info!').closest('div');
    expect(toast?.className).toContain('border-indigo-500');
  });
});
