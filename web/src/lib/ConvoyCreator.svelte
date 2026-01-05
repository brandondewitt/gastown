<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  interface CreateConvoyRequest {
    name: string;
    issue_ids: string[];
  }

  interface CreateConvoyResponse {
    id: string;
  }

  // State
  let convoyName: string = $state('');
  let issueInput: string = $state('');
  let issues: string[] = $state([]);
  let isLoading: boolean = $state(false);
  let error: string | null = $state(null);
  let success: string | null = $state(null);

  const dispatch = createEventDispatcher<{ created: { id: string; name: string } }>();

  // Add issue ID to the list
  function addIssue() {
    const trimmedIssue = issueInput.trim();
    if (trimmedIssue && !issues.includes(trimmedIssue)) {
      issues.push(trimmedIssue);
      issues = issues; // trigger reactivity
      issueInput = '';
    }
  }

  // Remove issue from the list
  function removeIssue(index: number) {
    issues.splice(index, 1);
    issues = issues; // trigger reactivity
  }

  // Handle Enter key in issue input
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      addIssue();
    }
  }

  // Submit form to create convoy
  async function handleSubmit(e: Event) {
    e.preventDefault();

    // Clear previous messages
    error = null;
    success = null;

    // Validate inputs
    if (!convoyName.trim()) {
      error = 'Convoy name is required';
      return;
    }

    if (issues.length === 0) {
      error = 'At least one issue ID is required';
      return;
    }

    isLoading = true;

    try {
      const request: CreateConvoyRequest = {
        name: convoyName.trim(),
        issue_ids: issues,
      };

      const response = await fetch('/api/v1/convoys', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error?.message || `HTTP error! status: ${response.status}`);
      }

      const data = (await response.json()) as CreateConvoyResponse;
      success = `Convoy created successfully with ID: ${data.id}`;

      // Dispatch event
      dispatch('created', { id: data.id, name: convoyName });

      // Reset form
      convoyName = '';
      issues = [];
      issueInput = '';
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create convoy';
    } finally {
      isLoading = false;
    }
  }

  function resetForm() {
    convoyName = '';
    issues = [];
    issueInput = '';
    error = null;
    success = null;
  }
</script>

<div class="convoy-creator">
  <h2>Create Convoy</h2>
  <form on:submit={handleSubmit}>
    <!-- Name Input -->
    <div class="form-group">
      <label for="convoy-name">Convoy Name</label>
      <input
        id="convoy-name"
        type="text"
        placeholder="e.g., Phase 1 Bug Fixes"
        bind:value={convoyName}
        disabled={isLoading}
      />
    </div>

    <!-- Issue Input -->
    <div class="form-group">
      <label for="issue-input">Issue IDs</label>
      <div class="issue-input-group">
        <input
          id="issue-input"
          type="text"
          placeholder="e.g., gt-om6zr (press Enter to add)"
          bind:value={issueInput}
          on:keydown={handleKeyDown}
          disabled={isLoading}
        />
        <button
          type="button"
          on:click={addIssue}
          disabled={isLoading || !issueInput.trim()}
          class="add-button"
        >
          Add
        </button>
      </div>
      {#if issues.length > 0}
        <div class="issues-list">
          {#each issues as issue, index}
            <div class="issue-tag">
              <span>{issue}</span>
              <button
                type="button"
                on:click={() => removeIssue(index)}
                disabled={isLoading}
                class="remove-button"
                title="Remove issue"
              >
                ×
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Error Message -->
    {#if error}
      <div class="alert alert-error">
        {error}
      </div>
    {/if}

    <!-- Success Message -->
    {#if success}
      <div class="alert alert-success">
        {success}
      </div>
    {/if}

    <!-- Form Actions -->
    <div class="form-actions">
      <button
        type="submit"
        disabled={isLoading || !convoyName.trim() || issues.length === 0}
        class="submit-button"
      >
        {isLoading ? 'Creating...' : 'Create Convoy'}
      </button>
      <button
        type="button"
        on:click={resetForm}
        disabled={isLoading}
        class="reset-button"
      >
        Clear
      </button>
    </div>
  </form>
</div>

<style>
  .convoy-creator {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--space-6);
  }

  .convoy-creator h2 {
    margin: 0 0 var(--space-4) 0;
    font-size: 1.125rem;
    font-weight: 600;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .form-group label {
    font-weight: 500;
    font-size: 0.875rem;
    color: var(--color-text);
  }

  .form-group input {
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }

  .form-group input:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
  }

  .form-group input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .issue-input-group {
    display: flex;
    gap: var(--space-2);
  }

  .issue-input-group input {
    flex: 1;
  }

  .add-button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-sm);
    font-weight: 500;
    cursor: pointer;
    font-size: 0.875rem;
    transition: background-color 0.2s;
  }

  .add-button:hover:not(:disabled) {
    background-color: var(--color-primary-hover, #2563eb);
  }

  .add-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .issues-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }

  .issue-tag {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-primary);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    color: var(--color-primary);
  }

  .remove-button {
    background: none;
    border: none;
    color: inherit;
    cursor: pointer;
    font-size: 1.25rem;
    padding: 0;
    line-height: 1;
    display: flex;
    align-items: center;
  }

  .remove-button:hover:not(:disabled) {
    color: var(--color-error, #ef4444);
  }

  .remove-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .alert {
    padding: var(--space-3);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
  }

  .alert-error {
    background-color: #fecaca;
    color: #7f1d1d;
    border: 1px solid #fca5a5;
  }

  .alert-success {
    background-color: #bbf7d0;
    color: #064e3b;
    border: 1px solid #6ee7b7;
  }

  .form-actions {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }

  .submit-button,
  .reset-button {
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-sm);
    font-weight: 500;
    cursor: pointer;
    font-size: 0.875rem;
    transition: all 0.2s;
  }

  .submit-button {
    background-color: var(--color-primary);
    color: white;
    flex: 1;
  }

  .submit-button:hover:not(:disabled) {
    background-color: var(--color-primary-hover, #2563eb);
  }

  .submit-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .reset-button {
    background-color: var(--color-surface-raised);
    color: var(--color-text-muted);
    border: 1px solid var(--color-border);
  }

  .reset-button:hover:not(:disabled) {
    background-color: var(--color-border);
  }

  .reset-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
