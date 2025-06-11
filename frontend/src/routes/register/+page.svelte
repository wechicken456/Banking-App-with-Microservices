<script lang="ts">
    import RegisterForm from '$lib/components/RegisterForm.svelte';
    import { redirectIfAuthenticated } from '$lib/utils/auth';

    // Redirect if already authenticated
    redirectIfAuthenticated();
</script>

<svelte:head>
    <title>Register - Banking App</title>
</svelte:head>

<div class="register-container">
    <div class="register-card">
        <div class="register-header">
            <h2 class="register-title">Create your account</h2>
            <p class="register-subtitle">
                Or
                <a href="/login" class="login-link">
                    sign in to your existing account
                </a>
            </p>
        </div>

        <RegisterForm />
    </div>
</div>

<style>
    @reference "../../app.css";
    
    .register-container {
        @apply min-h-screen flex items-center justify-center px-4 sm:px-6 lg:px-8;
    }

    .register-card {
        @apply max-w-md w-full space-y-8;
    }

    .register-header {
        @apply text-center;
    }

    .register-title {
        @apply text-3xl font-bold text-gray-900;
    }

    .register-subtitle {
        @apply mt-2 text-sm text-gray-600;
    }

    .login-link {
        @apply font-medium text-primary-600 hover:text-primary-500;
    }
</style>