import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { Eye, EyeOff, Mail, User, Lock, Check } from 'lucide-react';

export function RegisterPage() {
  const navigate = useNavigate();
  const { register, isLoading, user } = useAuthStore();
  
  // États du formulaire
  const [formData, setFormData] = useState({
    email: '',
    username: '',
    password: '',
    confirmPassword: '',
    acceptTerms: false
  });
  
  // États d'interface
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState<'success' | 'error'>('error');
  const [isRegistered, setIsRegistered] = useState(false);
  
  // États de validation
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [validations, setValidations] = useState<Record<string, boolean>>({});

  // Redirection si déjà connecté
  useEffect(() => {
    if (user) {
      navigate('/dashboard');
    }
  }, [user, navigate]);

  // Validation email
  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  // Validation nom d'utilisateur
  const validateUsername = (username: string): boolean => {
    const usernameRegex = /^[a-zA-Z0-9_]{3,30}$/;
    return usernameRegex.test(username);
  };

  // Calcul de la force du mot de passe
  const calculatePasswordStrength = (password: string) => {
    let score = 0;
    let feedback = [];

    if (password.length >= 8) score += 25;
    else feedback.push('Au moins 8 caractères');

    if (/[a-z]/.test(password)) score += 25;
    else feedback.push('Une minuscule');

    if (/[A-Z]/.test(password)) score += 25;
    else feedback.push('Une majuscule');

    if (/\d/.test(password)) score += 25;
    else feedback.push('Un chiffre');

    if (/[^a-zA-Z0-9]/.test(password)) score += 25;
    else feedback.push('Un caractère spécial');

    let strength = 'Très faible';
    let color = 'text-red-500';
    
    if (score >= 100) {
      strength = 'Très forte';
      color = 'text-green-500';
    } else if (score >= 75) {
      strength = 'Forte';
      color = 'text-green-400';
    } else if (score >= 50) {
      strength = 'Moyenne';
      color = 'text-yellow-500';
    } else if (score >= 25) {
      strength = 'Faible';
      color = 'text-orange-500';
    }

    return { score: Math.min(score, 100), strength, color, feedback };
  };

  // Validation d'un champ
  const validateField = (fieldName: string) => {
    const newErrors = { ...errors };
    const newValidations = { ...validations };

    switch (fieldName) {
      case 'email':
        if (!formData.email) {
          newErrors.email = 'L\'email est requis';
          newValidations.email = false;
        } else if (!validateEmail(formData.email)) {
          newErrors.email = 'Format d\'email invalide';
          newValidations.email = false;
        } else {
          delete newErrors.email;
          newValidations.email = true;
        }
        break;

      case 'username':
        if (!formData.username) {
          newErrors.username = 'Le nom d\'utilisateur est requis';
          newValidations.username = false;
        } else if (formData.username.length < 3) {
          newErrors.username = 'Au moins 3 caractères';
          newValidations.username = false;
        } else if (!validateUsername(formData.username)) {
          newErrors.username = 'Lettres, chiffres et _ uniquement';
          newValidations.username = false;
        } else {
          delete newErrors.username;
          newValidations.username = true;
        }
        break;

      case 'password':
        if (!formData.password) {
          newErrors.password = 'Le mot de passe est requis';
          newValidations.password = false;
        } else if (formData.password.length < 8) {
          newErrors.password = 'Au moins 8 caractères requis';
          newValidations.password = false;
        } else {
          delete newErrors.password;
          newValidations.password = true;
        }
        break;

      case 'confirmPassword':
        if (!formData.confirmPassword) {
          newErrors.confirmPassword = 'Confirmez votre mot de passe';
          newValidations.confirmPassword = false;
        } else if (formData.password !== formData.confirmPassword) {
          newErrors.confirmPassword = 'Les mots de passe ne correspondent pas';
          newValidations.confirmPassword = false;
        } else {
          delete newErrors.confirmPassword;
          newValidations.confirmPassword = true;
        }
        break;
    }

    setErrors(newErrors);
    setValidations(newValidations);
  };

  // Gestion des changements de formulaire
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));

    // Validation en temps réel
    if (name !== 'acceptTerms') {
      setTimeout(() => validateField(name), 300);
    }
  };

  // Vérification si le formulaire est valide
  const isFormValid = 
    validations.email &&
    validations.username &&
    validations.password &&
    validations.confirmPassword &&
    formData.acceptTerms;

  // Force du mot de passe
  const passwordStrength = formData.password ? calculatePasswordStrength(formData.password) : null;

  // Soumission du formulaire
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (isLoading || !isFormValid) return;
    
    setMessage('');
    
    // Validation finale
    Object.keys(formData).forEach(field => {
      if (field !== 'acceptTerms' && field !== 'confirmPassword') {
        validateField(field);
      }
    });
    validateField('confirmPassword');

    if (!isFormValid) {
      setMessage('Veuillez corriger les erreurs du formulaire');
      setMessageType('error');
      return;
    }
    
    try {
      await register(formData.email, formData.username, formData.password);
      setIsRegistered(true);
      setMessage('✅ Inscription réussie ! Redirection...');
      setMessageType('success');
      
      // Redirection vers le dashboard
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Erreur lors de l\'inscription');
      setMessageType('error');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-6 bg-gray-50">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-5xl font-extrabold tracking-tight bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent mb-2">
            🎶 Veza
          </h1>
          <h2 className="text-2xl font-bold text-gray-900">Créer un compte</h2>
          <p className="mt-2 text-sm text-gray-600">
            Rejoignez la communauté musicale Veza
          </p>
        </div>

        {/* Formulaire d'inscription */}
        <div className="bg-white rounded-lg shadow-lg p-8 space-y-6">
          <form onSubmit={handleRegister} className="space-y-4">
            {/* Email */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                Adresse email
              </label>
              <div className="relative">
                <Input
                  type="email"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleInputChange}
                  placeholder="votre@email.com"
                  required
                  disabled={isLoading}
                  className={`w-full pl-10 ${
                    errors.email ? 'border-red-500' : 
                    validations.email ? 'border-green-500' : ''
                  }`}
                />
                <Mail className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
              </div>
              {errors.email && (
                <p className="mt-1 text-sm text-red-600">{errors.email}</p>
              )}
            </div>

            {/* Nom d'utilisateur */}
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
                Nom d'utilisateur
              </label>
              <div className="relative">
                <Input
                  type="text"
                  id="username"
                  name="username"
                  value={formData.username}
                  onChange={handleInputChange}
                  placeholder="john_doe"
                  required
                  maxLength={30}
                  disabled={isLoading}
                  className={`w-full pl-10 ${
                    errors.username ? 'border-red-500' : 
                    validations.username ? 'border-green-500' : ''
                  }`}
                />
                <User className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
              </div>
              {errors.username && (
                <p className="mt-1 text-sm text-red-600">{errors.username}</p>
              )}
              {validations.username && (
                <p className="mt-1 text-sm text-green-600">✓ Nom d'utilisateur valide</p>
              )}
            </div>

            {/* Mot de passe */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                Mot de passe
              </label>
              <div className="relative">
                <Input
                  type={showPassword ? 'text' : 'password'}
                  id="password"
                  name="password"
                  value={formData.password}
                  onChange={handleInputChange}
                  placeholder="Minimum 8 caractères"
                  required
                  minLength={8}
                  disabled={isLoading}
                  className={`w-full pl-10 pr-10 ${
                    errors.password ? 'border-red-500' : 
                    validations.password ? 'border-green-500' : ''
                  }`}
                />
                <Lock className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-3 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {errors.password && (
                <p className="mt-1 text-sm text-red-600">{errors.password}</p>
              )}
              
              {/* Indicateur de force du mot de passe */}
              {formData.password && passwordStrength && (
                <div className="mt-2">
                  <div className="flex justify-between items-center mb-1">
                    <span className="text-xs text-gray-600">Force du mot de passe</span>
                    <span className={`text-xs ${passwordStrength.color}`}>
                      {passwordStrength.strength}
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className={`h-2 rounded-full transition-all duration-300 ${passwordStrength.color.replace('text-', 'bg-')}`}
                      style={{ width: `${passwordStrength.score}%` }}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Confirmation mot de passe */}
            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
                Confirmer le mot de passe
              </label>
              <div className="relative">
                <Input
                  type={showConfirmPassword ? 'text' : 'password'}
                  id="confirmPassword"
                  name="confirmPassword"
                  value={formData.confirmPassword}
                  onChange={handleInputChange}
                  placeholder="Retapez votre mot de passe"
                  required
                  disabled={isLoading}
                  className={`w-full pl-10 pr-10 ${
                    errors.confirmPassword ? 'border-red-500' : 
                    validations.confirmPassword ? 'border-green-500' : ''
                  }`}
                />
                <Check className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-3 text-gray-400 hover:text-gray-600"
                >
                  {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {errors.confirmPassword && (
                <p className="mt-1 text-sm text-red-600">{errors.confirmPassword}</p>
              )}
              {validations.confirmPassword && (
                <p className="mt-1 text-sm text-green-600">✓ Les mots de passe correspondent</p>
              )}
            </div>

            {/* Conditions d'utilisation */}
            <div className="flex items-start">
              <div className="flex items-center h-5">
                <input
                  id="acceptTerms"
                  name="acceptTerms"
                  type="checkbox"
                  checked={formData.acceptTerms}
                  onChange={handleInputChange}
                  required
                  disabled={isLoading}
                  className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 focus:ring-2"
                />
              </div>
              <div className="ml-3 text-sm">
                <label htmlFor="acceptTerms" className="text-gray-700">
                  J'accepte les{' '}
                  <Link to="/terms" className="text-blue-600 hover:text-blue-700 hover:underline">
                    conditions d'utilisation
                  </Link>
                  {' '}et la{' '}
                  <Link to="/privacy" className="text-blue-600 hover:text-blue-700 hover:underline">
                    politique de confidentialité
                  </Link>
                </label>
              </div>
            </div>

            {/* Bouton d'inscription */}
            <Button
              type="submit"
              disabled={!isFormValid || isLoading}
              className="w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors flex items-center justify-center gap-2 font-medium"
            >
              {isLoading ? (
                <>
                  <LoadingSpinner className="h-4 w-4" />
                  Création du compte...
                </>
              ) : (
                <>📝 Créer mon compte</>
              )}
            </Button>
          </form>

          {/* Message d'état */}
          {message && (
            <div className="text-center">
              <p className={`${messageType === 'success' ? 'text-green-600' : 'text-red-600'}`}>
                {message}
              </p>
            </div>
          )}

          {/* Navigation après inscription */}
          {isRegistered && (
            <div className="text-center space-y-3">
              <p className="text-green-600 font-medium">🎉 Bienvenue dans Veza !</p>
              <p className="text-sm text-gray-600">
                Vous allez être redirigé vers votre tableau de bord...
              </p>
            </div>
          )}
        </div>

        {/* Lien de connexion */}
        <div className="text-center">
          <p className="text-sm text-gray-600">
            Déjà un compte ?{' '}
            <Link to="/login" className="text-blue-600 hover:text-blue-700 font-medium hover:underline">
              Se connecter
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
} 