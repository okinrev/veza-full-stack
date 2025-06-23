import { z } from 'zod';

// Schéma pour le formulaire de connexion
export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'L\'email est requis')
    .email('Format email invalide'),
  password: z
    .string()
    .min(1, 'Le mot de passe est requis')
    .min(6, 'Le mot de passe doit contenir au moins 6 caractères'),
});

// Schéma pour le formulaire d'inscription
export const registerSchema = z.object({
  first_name: z
    .string()
    .min(1, 'Le prénom est requis')
    .min(2, 'Le prénom doit contenir au moins 2 caractères'),
  last_name: z
    .string()
    .min(1, 'Le nom est requis')
    .min(2, 'Le nom doit contenir au moins 2 caractères'),
  email: z
    .string()
    .min(1, 'L\'email est requis')
    .email('Format email invalide'),
  password: z
    .string()
    .min(1, 'Le mot de passe est requis')
    .min(6, 'Le mot de passe doit contenir au moins 6 caractères')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      'Le mot de passe doit contenir au moins une minuscule, une majuscule et un chiffre'
    ),
  confirmPassword: z
    .string()
    .min(1, 'La confirmation du mot de passe est requise'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Les mots de passe ne correspondent pas',
  path: ['confirmPassword'],
});

// Types dérivés des schémas
export type LoginFormData = z.infer<typeof loginSchema>;
export type RegisterFormData = z.infer<typeof registerSchema>;

// Schéma pour la réponse d'authentification
export const authResponseSchema = z.object({
  user: z.object({
    id: z.number(),
    email: z.string(),
    first_name: z.string(),
    last_name: z.string(),
    role: z.string().optional(),
  }),
  access_token: z.string(),
  refresh_token: z.string().optional(),
});

export type AuthResponse = z.infer<typeof authResponseSchema>; 