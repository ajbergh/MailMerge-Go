/**
 * ContactTable Component - Contact List Display with Selection and Filtering
 *
 * Selection identity is a stable contact ID (not the row index), so filtering,
 * searching, and sorting never corrupt which recipients are selected.
 */
import { useState, useMemo, useCallback } from 'react';
import { Contact } from '../types';
import { Users, Search, CheckSquare, Square, AlertTriangle } from 'lucide-react';

interface ContactTableProps {
  contacts: Contact[];
  /** Set of selected contact IDs (stable across filtering/sorting). */
  selectedIds?: Set<string>;
  onSelectionChange?: (selected: Set<string>) => void;
  /** List of duplicate email addresses to highlight. */
  duplicates?: string[];
}

export function ContactTable({ contacts, selectedIds, onSelectionChange, duplicates = [] }: ContactTableProps) {
  const [searchQuery, setSearchQuery] = useState('');

  const duplicateEmails = useMemo(() => new Set(duplicates.map((d) => d.toLowerCase())), [duplicates]);

  const filteredContacts = useMemo(() => {
    if (!searchQuery.trim()) {
      return contacts.map((contact, index) => ({ contact, originalIndex: index }));
    }
    const query = searchQuery.toLowerCase();
    return contacts
      .map((contact, index) => ({ contact, originalIndex: index }))
      .filter(({ contact }) =>
        contact.firstName?.toLowerCase().includes(query) ||
        contact.lastName?.toLowerCase().includes(query) ||
        contact.email?.toLowerCase().includes(query)
      );
  }, [contacts, searchQuery]);

  const selectionEnabled = selectedIds !== undefined && onSelectionChange !== undefined;

  const allFilteredSelected = useMemo(() => {
    if (!selectedIds || filteredContacts.length === 0) return false;
    return filteredContacts.every(({ contact }) => selectedIds.has(contact.id));
  }, [filteredContacts, selectedIds]);

  const toggleContact = useCallback((id: string) => {
    if (!onSelectionChange || !selectedIds) return;
    const next = new Set(selectedIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    onSelectionChange(next);
  }, [selectedIds, onSelectionChange]);

  const selectAll = useCallback(() => {
    if (!onSelectionChange) return;
    const next = new Set(selectedIds || []);
    filteredContacts.forEach(({ contact }) => next.add(contact.id));
    onSelectionChange(next);
  }, [filteredContacts, selectedIds, onSelectionChange]);

  const deselectAll = useCallback(() => {
    if (!onSelectionChange) return;
    const next = new Set(selectedIds || []);
    filteredContacts.forEach(({ contact }) => next.delete(contact.id));
    onSelectionChange(next);
  }, [filteredContacts, selectedIds, onSelectionChange]);

  if (contacts.length === 0) {
    return (
      <div className="card">
        <div className="card-header">
          <h2><Users size={20} />Contact Preview</h2>
        </div>
        <div className="card-body">
          <div className="empty-state">
            <Users className="empty-state-icon" />
            <h3>No Contacts Loaded</h3>
            <p>Import a CSV or Excel file to see your contacts here</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          <Users size={20} />
          Contact Preview ({contacts.length} contacts
          {selectionEnabled && selectedIds.size > 0 && `, ${selectedIds.size} selected`})
        </h2>
      </div>
      <div className="card-body">
        <div style={{ marginBottom: '1rem', display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <div style={{ position: 'relative', flex: 1, minWidth: '200px' }}>
            <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
            <label htmlFor="contact-search" className="sr-only">Search contacts by name or email</label>
            <input
              id="contact-search"
              type="text"
              placeholder="Search by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="form-input"
              style={{ paddingLeft: '2.25rem' }}
            />
          </div>
          {selectionEnabled && (
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button className="btn btn-sm btn-outline" onClick={selectAll} aria-pressed={allFilteredSelected} title="Select all visible contacts">
                <CheckSquare size={14} />Select All
              </button>
              <button className="btn btn-sm btn-outline" onClick={deselectAll} title="Deselect all visible contacts">
                <Square size={14} />Deselect All
              </button>
            </div>
          )}
        </div>

        {duplicates.length > 0 && (
          <div className="alert alert-warning" style={{ marginBottom: '1rem' }}>
            <AlertTriangle size={16} />
            <span>Found {duplicates.length} duplicate email address(es). Highlighted rows indicate duplicates.</span>
          </div>
        )}

        <div className="table-container scrollable-table">
          <table className="table">
            <thead>
              <tr>
                {selectionEnabled && <th style={{ width: '40px' }} scope="col"><span className="sr-only">Select</span></th>}
                <th scope="col">#</th>
                <th scope="col">First Name</th>
                <th scope="col">Last Name</th>
                <th scope="col">Email</th>
              </tr>
            </thead>
            <tbody>
              {filteredContacts.map(({ contact, originalIndex }) => {
                const isDuplicate = duplicateEmails.has(contact.email.toLowerCase());
                const isSelected = selectionEnabled && selectedIds.has(contact.id);
                return (
                  <tr
                    key={contact.id || originalIndex}
                    style={{ backgroundColor: isDuplicate ? 'var(--warning-light)' : undefined, cursor: selectionEnabled ? 'pointer' : undefined }}
                    onClick={() => selectionEnabled && toggleContact(contact.id)}
                  >
                    {selectionEnabled && (
                      <td style={{ textAlign: 'center' }} onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => toggleContact(contact.id)}
                          aria-label={`Select ${contact.email}`}
                          style={{ cursor: 'pointer' }}
                        />
                      </td>
                    )}
                    <td style={{ color: 'var(--text-muted)' }}>{originalIndex + 1}</td>
                    <td>{contact.firstName || <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>—</span>}</td>
                    <td>{contact.lastName || <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>—</span>}</td>
                    <td>
                      {contact.email}
                      {isDuplicate && (
                        <span title="Duplicate email address">
                          <AlertTriangle size={14} style={{ marginLeft: '0.5rem', color: 'var(--warning)', verticalAlign: 'middle' }} />
                        </span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {filteredContacts.length === 0 && searchQuery && (
          <div className="empty-state" style={{ padding: '2rem' }}>
            <Search className="empty-state-icon" style={{ width: 32, height: 32 }} />
            <p>No contacts match &ldquo;{searchQuery}&rdquo;</p>
          </div>
        )}
      </div>
    </div>
  );
}
