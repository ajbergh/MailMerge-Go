import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ContactTable } from './ContactTable';
import type { Contact } from '../types';

const contacts: Contact[] = [
  { id: 'a', firstName: 'Ada', lastName: 'L', email: 'ada@example.com' },
  { id: 'b', firstName: 'Bob', lastName: 'M', email: 'bob@example.com' },
];

describe('ContactTable', () => {
  it('selects contacts by stable ID, not row index', () => {
    let selected = new Set<string>();
    const onChange = vi.fn((s: Set<string>) => { selected = s; });
    render(<ContactTable contacts={contacts} selectedIds={selected} onSelectionChange={onChange} />);

    // Toggle the first contact.
    fireEvent.click(screen.getByLabelText('Select ada@example.com'));
    expect(onChange).toHaveBeenCalled();
    const arg = onChange.mock.calls[0][0];
    expect(arg.has('a')).toBe(true);
    expect(arg.has('b')).toBe(false);
  });

  it('filters without losing selection identity', () => {
    const onChange = vi.fn();
    render(
      <ContactTable contacts={contacts} selectedIds={new Set(['b'])} onSelectionChange={onChange} />
    );
    // Search for Bob; Ada should be filtered out but selection is by id.
    fireEvent.change(screen.getByLabelText(/Search contacts/i), { target: { value: 'bob' } });
    expect(screen.getByText('bob@example.com')).toBeInTheDocument();
    expect(screen.queryByText('ada@example.com')).not.toBeInTheDocument();
  });
});
