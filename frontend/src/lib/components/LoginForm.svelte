<script lang="ts">
    import Button from '$lib/components/ui/Button.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import { validateEmail, validatePassword } from '$lib/utils/validate';
    import { authStore } from '$lib/stores/auth.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';

    let email = $state('');
    let password = $state('');
    let emailError = $state<string | null>(null);
    let passwordError = $state<string | null>(null);

    const isLoading = $derived(authStore.isLoading);

    function validateForm() {
        emailError = validateEmail(email);
        passwordError = validatePassword(password);
        return !emailError && !passwordError;
    }

    async function handleSubmit(e: SubmitEvent) {
        e.preventDefault();
        if (!validateForm()) return;

        const result = await authStore.login(email, password);
        
        if (result.success) {
            toastStore.success('Login successful!');
        } else {
            toastStore.error(result.error || 'Login failed');
        }
    }
</script>

<form onsubmit={handleSubmit} class="form">
    <div class="form-field">
        <Input
            label="Email"
            type="email"
            placeholder="Enter your email"
            bind:value={email}
            error={emailError}
            required
        />
    </div>

    <div class="form-field">
        <Input
            label="Password"
            type="password"
            placeholder="Enter your password"
            bind:value={password}
            error={passwordError}
            required
        />
    </div>

    <Button
        type="submit"
        variant="primary"
        size="lg"
        loading={isLoading}
        disabled={isLoading}
    >
        {isLoading ? 'Signing in...' : 'Sign in'}
    </Button>
</form>

<style>
    @reference "../../app.css";

    .form {
        @apply space-y-6;
    }

    .form-field {
        @apply w-full;
    }
</style>