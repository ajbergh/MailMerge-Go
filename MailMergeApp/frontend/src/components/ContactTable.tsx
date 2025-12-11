/**
 * ContactTable Component - Contact List Display with Selection and Filtering
 * 
 * This component displays a scrollable table preview of imported contacts.
 * It shows the first name, last name, and email address for each contact
 * with row numbering for easy reference.
 * 
 * Phase 1 Updates (v1.2):
 * - Added search/filter functionality to filter by name or email
 * - Added checkbox selection for choosing specific contacts to send to
 * - Select All / Deselect All buttons
 * - Visual indication of selected contacts count
 * - Duplicate email highlighting
 * 
 * Features:
 * - Empty state with helpful message when no contacts loaded
 * - Scrollable table for large contact lists
 * - Graceful handling of missing name fields (shows dash)
 * - Row numbering for easy reference
 * - Search filtering (name or email)
 * - Checkbox selection for targeted sends
 */
import { useState, useMemo, useCallback } from 'react';
import { Contact } from '../types';
import { Users, Search, CheckSquare, Square, AlertTriangle } from 'lucide-react';

/**
 * Props for the ContactTable component
 */
interface ContactTableProps {
  /** Array of contacts to display in the table */
  contacts: Contact[];
  /** Set of selected contact indices (for targeted sending) */
  selectedIndices?: Set<number>;
  /** Callback when selection changes */
  onSelectionChange?: (selected: Set<number>) => void;
  /** List of duplicate email addresses to highlight */
  duplicates?: string[];
}

/**
 * ContactTable - Renders a table preview of loaded contacts with selection and filtering
 * 
 * Shows either:
 * - Empty state with prompt to import contacts
 * - Scrollable table with search, selection, and contact data
 */
export function ContactTable({ 
  contacts, 
  selectedIndices, 
  onSelectionChange,
  duplicates = []
}: ContactTableProps) {
  // Search/filter state
  const [searchQuery, setSearchQuery] = useState('');

  // Create a set of duplicate emails (lowercase for comparison)
  const duplicateEmails = useMemo(() => 
    new Set(duplicates.map(d => d.toLowerCase())),
    [duplicates]
  );

  // Filter contacts based on search query
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

  // Check if all filtered contacts are selected
  const allFilteredSelected = useMemo(() => {
    if (!selectedIndices || filteredContacts.length === 0) return false;
    return filteredContacts.every(({ originalIndex }) => selectedIndices.has(originalIndex));
  }, [filteredContacts, selectedIndices]);

  // Toggle single contact selection
  const toggleContact = useCallback((originalIndex: number) => {
    if (!onSelectionChange || !selectedIndices) return;
    const newSelection = new Set(selectedIndices);
    if (newSelection.has(originalIndex)) {
      newSelection.delete(originalIndex);
    } else {
      newSelection.add(originalIndex);
    }
    onSelectionChange(newSelection);
  }, [selectedIndices, onSelectionChange]);

  // Select all filtered contacts
  const selectAll = useCallback(() => {
    if (!onSelectionChange) return;
    const newSelection = new Set(selectedIndices || []);
    filteredContacts.forEach(({ originalIndex }) => newSelection.add(originalIndex));
    onSelectionChange(newSelection);
  }, [filteredContacts, selectedIndices, onSelectionChange]);

  // Deselect all filtered contacts
  const deselectAll = useCallback(() => {
    if (!onSelectionChange) return;
    const newSelection = new Set(selectedIndices || []);
    filteredContacts.forEach(({ originalIndex }) => newSelection.delete(originalIndex));
    onSelectionChange(newSelection);
  }, [filteredContacts, selectedIndices, onSelectionChange]);

  // Check if selection mode is enabled
  const selectionEnabled = selectedIndices !== undefined && onSelectionChange !== undefined;

  if (contacts.length === 0) {
    return (
      <div className="card">
        <div className="card-header">
          <h2>
            <Users size={20} />
            Contact Preview
          </h2>
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
          {selectionEnabled && selectedIndices.size > 0 && 
            `, ${selectedIndices.size} selected`
          })
        </h2>
      </div>
      <div className="card-body">
        {/* Search and selection controls */}
        <div style={{ marginBottom: '1rem', display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
          {/* Search input */}
          <div style={{ position: 'relative', flex: 1, minWidth: '200px' }}>
            <Search 
              size={16} 
              style={{ 
                position: 'absolute', 
                left: '0.75rem', 
                top: '50%', 
                transform: 'translateY(-50%)',
                color: 'var(--text-muted)'
              }} 
            />
            <input
              type="text"
              placeholder="Search by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="form-input"
              style={{ paddingLeft: '2.25rem' }}
            />
          </div>

          {/* Selection buttons */}
          {selectionEnabled && (
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button 
                className="btn btn-sm btn-outline" 
                onClick={selectAll}
                title="Select all visible contacts"
              >
                <CheckSquare size={14} />
                Select All
              </button>
              <button 
                className="btn btn-sm btn-outline" 
                onClick={deselectAll}
                title="Deselect all visible contacts"
              >
                <Square size={14} />
                Deselect All
              </button>
            </div>
          )}
        </div>

        {/* Duplicate warning */}
        {duplicates.length > 0 && (
          <div className="alert alert-warning" style={{ marginBottom: '1rem' }}>
            <AlertTriangle size={16} />
            <span>Found {duplicates.length} duplicate email address(es). Highlighted rows indicate duplicates.</span>
          </div>
        )}

        {/* Contact table */}
        <div className="table-container scrollable-table">
          <table className="table">
            <thead>
              <tr>
                {selectionEnabled && <th style={{ width: '40px' }}></th>}
                <th>#</th>
                <th>First Name</th>
                <th>Last Name</th>
                <th>Email</th>
              </tr>
            </thead>
            <tbody>
              {filteredContacts.map(({ contact, originalIndex }) => {
                const isDuplicate = duplicateEmails.has(contact.email.toLowerCase());
                const isSelected = selectionEnabled && selectedIndices.has(originalIndex);
                
                return (
                  <tr 
                    key={originalIndex}
                    style={{
                      backgroundColor: isDuplicate ? 'var(--warning-light)' : undefined,
                      cursor: selectionEnabled ? 'pointer' : undefined
                    }}
                    onClick={() => selectionEnabled && toggleContact(originalIndex)}
                  >
                    {selectionEnabled && (
                      <td style={{ textAlign: 'center' }} onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => toggleContact(originalIndex)}
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
                          <AlertTriangle 
                            size={14} 
                            style={{ marginLeft: '0.5rem', color: 'var(--warning)', verticalAlign: 'middle' }} 
                          />
                        </span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {/* Show message when search yields no results */}
        {filteredContacts.length === 0 && searchQuery && (
          <div className="empty-state" style={{ padding: '2rem' }}>
            <Search className="empty-state-icon" style={{ width: 32, height: 32 }} />
            <p>No contacts match "{searchQuery}"</p>
          </div>
        )}
      </div>
    </div>
  );
}
