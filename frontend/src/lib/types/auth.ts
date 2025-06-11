export interface User {
    id: string;
    email: string;
}

export interface LoginCredentials {
    email: string;
    password: string;
}

export interface RegisterCredentials extends LoginCredentials {
    confirmPassword: string;
}

export interface LoginResponse {
    accessToken: string;
    refreshToken: string;
    fingerprint: string;
    user: User;
}

export interface ApiError {
    message: string;
    code?: string;
}


