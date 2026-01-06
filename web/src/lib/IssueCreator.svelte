<script lang="ts">
  interface Props {
    onCreated?: () => void;
  }

  let { onCreated }: Props = $props();

  interface Rig {
    name: string;
    path: string;
  }

  // State
  let title: string = $state('');
  let description: string = $state('');
  let issueType: string = $state('task');
  let priority: string = $state('2');
  let selectedRig: string = $state('');
  let rigs: Rig[] = $state([]);
  let isLoading: boolean = $state(false);
  let rigsLoading: boolean = $state(true);
  let error: string | null = $state(null);
  let success: string | null = $state(null);

  // Fetch rigs on mount
  $effect(() => {
    fetchRigs();
  });

  async function fetchRigs() {
    try {
      const response = await fetch('/api/v1/rigs');
      const data = await response.json();
      if (data.success && data.data) {
        rigs = data.data;
      }
    } catch {
      // Silently ignore - rig selection is optional
    } finally {
      rigsLoading = false;
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();

    // Clear previous messages
    error = null;
    success = null;

    // Validate inputs
    if (!title.trim()) {
      error = 'Title is required';
      return;
    }

    isLoading = true;

    try {
      const request: Record<string, string> = {
        title: title.trim(),
        type: issueType,
        priority: priority,
      };

      if (description.trim()) {
        request.description = description.trim();
      }

      if (selectedRig) {
        request.rig = selectedRig;
      }

      const response = await fetch('/api/v1/beads', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      const data = await response.json();

      if (!response.ok || !data.success) {
        throw new Error(data.error?.message || `HTTP error! status: ${response.status}`);
      }

      success = `Issue created: ${data.id}`;

      // Reset form
      title = '';
      description = '';
      issueType = 'task';
      priority = '2';
      selectedRig = '';

      // Notify parent
      if (onCreated) {
        setTimeout(() => onCreated(), 1500);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create issue';
    } finally {
      isLoading = false;
    }
  }

  function resetForm() {
    title = '';
    description = '';
    issueType = 'task';
    priority = '2';
    selectedRig = '';
    error = null;
    success = null;
  }
</script>

<div class="issue-creator">
  <h2>Create Issue</h2>
  <form onsubmit={handleSubmit}>
    <!-- Title Input -->
    <div class="form-group">
      <label for="issue-title">Title *</label>
      <input
        id="issue-title"
        type="text"
        placeholder="Brief description of the issue"
        bind:value={title}
        disabled={isLoading}
      />
    </div>

    <!-- Type and Priority Row -->
    <div class="form-row">
      <div class="form-group">
        <label for="issue-type">Type</label>
        <select id="issue-type" bind:value={issueType} disabled={isLoading}>
          <option value="task">Task</option>
          <option value="bug">Bug</option>
          <option value="feature">Feature</option>
          <option value="epic">Epic</option>
          <option value="chore">Chore</option>
        </select>
      </div>

      <div class="form-group">
        <label for="issue-priority">Priority</label>
        <select id="issue-priority" bind:value={priority} disabled={isLoading}>
          <option value="0">P0 - Critical</option>
          <option value="1">P1 - High</option>
          <option value="2">P2 - Medium</option>
          <option value="3">P3 - Low</option>
          <option value="4">P4 - Minimal</option>
        </select>
      </div>
    </div>

    <!-- Rig Selection -->
    <div class="form-group">
      <label for="issue-rig">Rig (optional)</label>
      <select id="issue-rig" bind:value={selectedRig} disabled={isLoading || rigsLoading}>
        <option value="">Default (current rig)</option>
        {#each rigs as rig}
          <option value={rig.name}>{rig.name}</option>
        {/each}
      </select>
    </div>

    <!-- Description -->
    <div class="form-group">
      <label for="issue-description">Description (optional)</label>
      <textarea
        id="issue-description"
        placeholder="Additional details about the issue..."
        bind:value={description}
        disabled={isLoading}
        rows="3"
      ></textarea>
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
        disabled={isLoading || !title.trim()}
        class="submit-button"
      >
        {isLoading ? 'Creating...' : 'Create Issue'}
      </button>
      <button
        type="button"
        onclick={resetForm}
        disabled={isLoading}
        class="reset-button"
      >
        Clear
      </button>
    </div>
  </form>
</div>

<style>
  .issue-creator {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--space-6);
  }

  .issue-creator h2 {
    margin: 0 0 var(--space-4) 0;
    font-size: 1.125rem;
    font-weight: 600;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
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

  .form-group input,
  .form-group select,
  .form-group textarea {
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }

  .form-group textarea {
    resize: vertical;
    min-height: 80px;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
  }

  .form-group input:disabled,
  .form-group select:disabled,
  .form-group textarea:disabled {
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
