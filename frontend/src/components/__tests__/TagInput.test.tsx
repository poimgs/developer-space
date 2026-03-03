import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import TagInput from '../TagInput';

describe('TagInput', () => {
  it('renders existing tags as pills', () => {
    render(<TagInput value={['react', 'typescript']} onChange={vi.fn()} />);
    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByText('typescript')).toBeInTheDocument();
  });

  it('adds a tag on Enter', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={[]} onChange={onChange} />);
    const input = screen.getByPlaceholderText('Add a tag…');
    await user.type(input, 'react{Enter}');
    expect(onChange).toHaveBeenCalledWith(['react']);
  });

  it('adds a tag on comma', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={[]} onChange={onChange} />);
    const input = screen.getByPlaceholderText('Add a tag…');
    await user.type(input, 'golang,');
    expect(onChange).toHaveBeenCalledWith(['golang']);
  });

  it('trims and lowercases tags', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={[]} onChange={onChange} />);
    const input = screen.getByPlaceholderText('Add a tag…');
    await user.type(input, '  TypeScript  {Enter}');
    expect(onChange).toHaveBeenCalledWith(['typescript']);
  });

  it('removes a tag when × is clicked', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={['react', 'go']} onChange={onChange} />);
    await user.click(screen.getByLabelText('Remove react'));
    expect(onChange).toHaveBeenCalledWith(['go']);
  });

  it('prevents duplicate tags', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={['react']} onChange={onChange} />);
    const input = screen.getByPlaceholderText('Add a tag…');
    await user.type(input, 'react{Enter}');
    expect(onChange).not.toHaveBeenCalled();
  });

  it('ignores empty input on Enter', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={[]} onChange={onChange} />);
    const input = screen.getByPlaceholderText('Add a tag…');
    await user.type(input, '   {Enter}');
    expect(onChange).not.toHaveBeenCalled();
  });

  it('hides input when max tags reached', () => {
    const tags = ['a', 'b', 'c'];
    render(<TagInput value={tags} onChange={vi.fn()} max={3} />);
    expect(screen.queryByPlaceholderText('Add a tag…')).toBeNull();
    expect(screen.getByText('Maximum of 3 tags reached')).toBeInTheDocument();
  });

  it('removes last tag on Backspace with empty input', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagInput value={['react', 'go']} onChange={onChange} />);
    const input = screen.getByPlaceholderText('Add a tag…');
    await user.click(input);
    await user.keyboard('{Backspace}');
    expect(onChange).toHaveBeenCalledWith(['react']);
  });

  it('uses custom placeholder', () => {
    render(<TagInput value={[]} onChange={vi.fn()} placeholder="Type skill…" />);
    expect(screen.getByPlaceholderText('Type skill…')).toBeInTheDocument();
  });
});
