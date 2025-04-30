export const passwordValidation = {
  required: 'Password is required',
};

export const passwordSignupValidation = {
  required: 'Password is required',
  minLength: {
    value: 6,
    message: 'Password must be at least 6 characters',
  },
  validate: {
    hasUppercase: (value: string) =>
      /[A-Z]/.test(value) ||
      'Password must contain at least one uppercase letter',
    hasLowercase: (value: string) =>
      /[a-z]/.test(value) ||
      'Password must contain at least one lowercase letter',
    hasNumber: (value: string) =>
      /[0-9]/.test(value) ||
      'Password must contain at least one number',
    hasSpecialChar: (value: string) =>
      /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(value) ||
      'Password must contain at least one special character',
  },
};

export const emailValidation = {
  required: 'Email is required',
  pattern: {
    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
    message: 'Please enter a valid email address',
  },
};
