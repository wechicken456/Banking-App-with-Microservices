<script lang="ts">
    import Button from '$lib/components/ui/Button.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import { validateEmail, validatePassword, validateConfirmPassword } from '$lib/utils/validate';
    import { authStore } from '$lib/stores/auth.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import { goto } from '$app/navigation';

    let email = $state('');
    let password = $state('');
    let confirmPassword = $state('');
    let emailError = $state<string | null>(null);
    let passwordError = $state<string | null>(null);
    let confirmPasswordError = $state<string | null>(null);

    const isLoading = $derived(authStore.isLoading);

    function validateForm() {
        emailError = validateEmail(email);
        passwordError = validatePassword(password);
        confirmPasswordError = validateConfirmPassword(password, confirmPassword);
        return !emailError && !passwordError && !confirmPasswordError;
    }

    async function handleSubmit(e: SubmitEvent) {
        e.preventDefault();
        
        if (!validateForm()) return;

        const result = await authStore.register(email, password, confirmPassword);
        
        if (result.success) {
            toastStore.success('Registration successful! Please sign in.');
            goto('/login');
        } else {
            toastStore.error(result.error || 'Registration failed');
        }
    }
</script>

<div class="register-form">
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

        <div class="form-field">
            <Input
                label="Confirm Password"
                type="password"
                placeholder="Confirm your password"
                bind:value={confirmPassword}
                error={confirmPasswordError}
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
            {isLoading ? 'Creating account...' : 'Create account'}
        </Button>
    </form>
</div>

<style>

    @reference "../../app.css";

    .register-form {
        @apply bg-white py-8 px-6 shadow rounded-lg;
    }

    .form {
        @apply space-y-6;
    }

    .form-field {
        @apply w-full;
    }
</style>