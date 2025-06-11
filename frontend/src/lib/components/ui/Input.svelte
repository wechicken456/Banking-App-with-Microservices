<script lang="ts">
    interface Props {
        label?: string;
        type?: string;
        placeholder?: string;
        value?: string;
        error?: string | null;
        required?: boolean;
        disabled?: boolean;
        id?: string;
    }

    let {
        label,
        type = 'text',
        placeholder,
        value = $bindable(''),
        error,
        required = false,
        disabled = false,
        id
    }: Props = $props();

    const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
</script>

<div class="input-container">
    {#if label}
        <label for={inputId} class="input-label">
            {label}
            {#if required}
                <span class="required">*</span>
            {/if}
        </label>
    {/if}
    <input
        id={inputId}
        {type}
        {placeholder}
        {required}
        {disabled}
        bind:value
        class="input {error ? 'input-error' : ''}"
    />
    {#if error}
        <p class="error-message">{error}</p>
    {/if}
</div>

<style>
    @reference "../../../app.css";

    .input-container {
        @apply w-full;
    }

    .input-label {
        @apply block text-sm font-medium text-gray-700 mb-1;
    }

    .required {
        @apply text-red-500;
    }

    .input {
        @apply w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-primary-500 focus:border-primary-500 disabled:bg-gray-50 disabled:text-gray-500;
    }

    .input-error {
        @apply border-red-500 focus:ring-red-500 focus:border-red-500;
    }

    .error-message {
        @apply mt-1 text-sm text-red-600;
    }
</style>