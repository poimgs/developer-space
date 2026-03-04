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

  describe('suggestions', () => {
    const suggestions = ['react', 'redux', 'rust', 'go', 'graphql'];

    it('shows filtered suggestions dropdown on input', async () => {
      const user = userEvent.setup();
      render(<TagInput value={[]} onChange={vi.fn()} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      expect(screen.getByRole('listbox')).toBeInTheDocument();
      expect(screen.getByText('react')).toBeInTheDocument();
      expect(screen.getByText('redux')).toBeInTheDocument();
      expect(screen.queryByText('go')).not.toBeInTheDocument();
    });

    it('filters out already-selected tags from suggestions', async () => {
      const user = userEvent.setup();
      render(<TagInput value={['react']} onChange={vi.fn()} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      expect(screen.queryByRole('option', { name: 'react' })).not.toBeInTheDocument();
      expect(screen.getByText('redux')).toBeInTheDocument();
    });

    it('navigates suggestions with ArrowDown/ArrowUp', async () => {
      const user = userEvent.setup();
      render(<TagInput value={[]} onChange={vi.fn()} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      await user.keyboard('{ArrowDown}');
      expect(screen.getByRole('option', { name: 'react' })).toHaveAttribute('aria-selected', 'true');
      await user.keyboard('{ArrowDown}');
      expect(screen.getByRole('option', { name: 'redux' })).toHaveAttribute('aria-selected', 'true');
      await user.keyboard('{ArrowUp}');
      expect(screen.getByRole('option', { name: 'react' })).toHaveAttribute('aria-selected', 'true');
    });

    it('selects highlighted suggestion on Enter', async () => {
      const user = userEvent.setup();
      const onChange = vi.fn();
      render(<TagInput value={[]} onChange={onChange} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      await user.keyboard('{ArrowDown}{ArrowDown}{Enter}');
      expect(onChange).toHaveBeenCalledWith(['redux']);
    });

    it('selects first suggestion on Tab when dropdown is visible', async () => {
      const user = userEvent.setup();
      const onChange = vi.fn();
      render(<TagInput value={[]} onChange={onChange} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      await user.keyboard('{Tab}');
      expect(onChange).toHaveBeenCalledWith(['react']);
    });

    it('closes suggestions on Escape', async () => {
      const user = userEvent.setup();
      render(<TagInput value={[]} onChange={vi.fn()} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      expect(screen.getByRole('listbox')).toBeInTheDocument();
      await user.keyboard('{Escape}');
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });

    it('adds suggestion on mousedown', async () => {
      const user = userEvent.setup();
      const onChange = vi.fn();
      render(<TagInput value={[]} onChange={onChange} suggestions={suggestions} />);
      const input = screen.getByPlaceholderText('Add a tag…');
      await user.type(input, 're');
      // Click on the suggestion item
      const option = screen.getByText('react');
      await user.click(option);
      expect(onChange).toHaveBeenCalledWith(['react']);
    });

    it('has combobox role and aria attributes', async () => {
      const user = userEvent.setup();
      render(<TagInput value={[]} onChange={vi.fn()} suggestions={suggestions} />);
      const input = screen.getByRole('combobox');
      expect(input).toHaveAttribute('aria-expanded', 'false');
      await user.type(input, 're');
      expect(input).toHaveAttribute('aria-expanded', 'true');
      expect(input).toHaveAttribute('aria-controls', 'tag-suggestions-listbox');
    });
  });
});
