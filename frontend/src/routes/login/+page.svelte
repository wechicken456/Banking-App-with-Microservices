<script lang="ts">
    import LoginForm from '$lib/components/LoginForm.svelte';
    import { authStore } from '$lib/stores/auth.svelte';
    import { redirectIfAuthenticated } from '$lib/utils/auth';

    // Redirect if already authenticated
    redirectIfAuthenticated();
</script>

<svelte:head>
    <title>Login - Banking App</title>
</svelte:head>

<div class="login-container">
    <div class="login-card">
        <div class="login-header">
            <h2 class="login-title">Sign in to your account</h2>
            <p class="login-subtitle">
                Or
                <a href="/register" class="register-link">
                    create a new account
                </a>
            </p>
        </div>

        <div class="login-form">
            <LoginForm />
        </div>
    </div>
</div>

<style>
    @reference "../../app.css";
    .login-container {
        @apply min-h-screen flex items-center justify-center px-4 sm:px-6 lg:px-8;
    }

    .login-card {
        @apply max-w-md w-full space-y-8;
    }

    .login-header {
        @apply text-center;
    }

    .login-title {
        @apply text-3xl font-bold text-gray-900;
    }

    .login-subtitle {
        @apply mt-2 text-sm text-gray-600;
    }

    .register-link {
        @apply font-medium text-primary-600 hover:text-primary-500;
    }

    .login-form {
        @apply bg-white py-8 px-6 shadow rounded-lg;
    }
</style>